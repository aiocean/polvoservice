package server

import (
	"context"
	"fmt"

	"github.com/google/wire"
	fieldmask_utils "github.com/mennanov/fieldmask-utils"
	"github.com/nguyenvanduocit/toCamelCase"
	"github.com/nguyenvanduocit/toSnakeCase"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	polvo_v1 "pkg.aiocean.dev/polvogo/aiocean/polvo/v1"
	"pkg.aiocean.dev/polvoservice/internal/repository"
	"pkg.aiocean.dev/serviceutil/handler"
)

var WireSet = wire.NewSet(
	NewServer,
	wire.Bind(new(handler.ServiceServer), new(*Server)),
)

var defaultVersions = map[string]string{
	"any": "any",
}

type Server struct {
	logger *zap.Logger
	repo   repository.Repository
	polvo_v1.UnimplementedPolvoServiceServer
}

func NewServer(logger *zap.Logger, repo repository.Repository) *Server {
	return &Server{
		logger: logger,
		repo:   repo,
	}
}


func (s *Server) Register(grpcServer *grpc.Server) {
	polvo_v1.RegisterPolvoServiceServer(grpcServer, s)
}

func (s *Server) CreatePackage(request *polvo_v1.CreatePackageRequest, stream polvo_v1.PolvoService_CreatePackageServer) error {

	isPackageExist, err := s.repo.IsPackageExists(stream.Context(), request.GetPackage().GetName())
	if err != nil {
		return errors.New("Can not check of package")
	}

	if isPackageExist {
		return status.Error(codes.AlreadyExists, "package is already exists")
	}

	savedPackage, err := s.repo.CreatePackage(stream.Context(), request.GetPackage())
	if  err != nil {
		return errors.Wrap(err, "failed to get package")
	}

	response := &polvo_v1.CreatePackageResponse{
		Package: savedPackage,
	}

	if err := stream.Send(response); err != nil {
		return errors.Wrap(err, "failed to send response")
	}

	return nil
}

func (s *Server) DeletePackage(request *polvo_v1.DeletePackageRequest, stream polvo_v1.PolvoService_DeletePackageServer) error {

	packageName := parsePackageOrn(request.GetOrn())

	isPackageExists, err := s.repo.IsPackageExists(stream.Context(), packageName)
	if err != nil {
		return err
	}

	if !isPackageExists {
		if err := stream.Send(&polvo_v1.DeletePackageResponse{
			Message: "Package is not exists",
		}); err != nil {
			return err
		}
	}

	if err := s.repo.DeletePackage(stream.Context(), packageName); err != nil {
		return status.Errorf(codes.Internal, "failed to delete package", err.Error())
	}

	if err := stream.Send(&polvo_v1.DeletePackageResponse{
		Message: "Package and its version are deleted",
	}); err != nil {
		return err
	}

	return nil
}

func (s *Server) UpdatePackage(ctx context.Context, request *polvo_v1.UpdatePackageRequest) (*polvo_v1.UpdatePackageResponse, error) {
	packageName := parsePackageOrn(request.GetOrn())

	isPackageExists, err := s.repo.IsPackageExists(ctx, packageName)
	if err != nil {
		return nil, err
	}

	if !isPackageExists {
		return nil, status.Error(codes.NotFound, "package not found")
	}

	request.GetFieldMask().Normalize()
	if !request.GetFieldMask().IsValid(request) {
		return nil, errors.New("update mask is invalid")
	}

	filteredRequest := map[string]interface{}{}

	mask, err := fieldmask_utils.MaskFromProtoFieldMask(request.GetFieldMask(), toSnakeCase.ToSnakeCase)
	if err != nil {
		return nil, err
	}

	if err := fieldmask_utils.StructToMap(mask, request, filteredRequest); err != nil {
		return nil, err
	}

	updateFields := filteredRequest[string(request.GetPackage().ProtoReflect().Descriptor().Name())].(map[string]interface{})

	savedPackage, err := s.repo.UpdatePackage(ctx, packageName,updateFields)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get package")
	}

	return &polvo_v1.UpdatePackageResponse{
		Package: savedPackage,
	}, nil
}

func (s *Server)UpdateVersion(request *polvo_v1.UpdateVersionRequest, stream polvo_v1.PolvoService_UpdateVersionServer) error {

	packageName, versionName := parseVersionOrn(request.GetOrn())

	isVersionExists, err := s.repo.IsVersionExists(stream.Context(), packageName, versionName)
	if err != nil {
		return err
	}

	if !isVersionExists {
		return status.Error(codes.NotFound, "version not found")
	}

	request.GetFieldMask().Normalize()
	if !request.GetFieldMask().IsValid(request) {
		return errors.New("update mask is invalid")
	}

	mask, err := fieldmask_utils.MaskFromProtoFieldMask(request.GetFieldMask(), toCamelCase.ToCamelCase)
	if err != nil {
		return err
	}

	filteredRequest := make(map[string]interface{})

	if err := fieldmask_utils.StructToMap(mask, request, filteredRequest); err != nil {
		return err
	}

	var updateFields map[string]interface{}

	if fields, ok := filteredRequest["Version"]; ok {
		updateFields = fields.(map[string]interface{})
	} else {
		return status.Error(codes.InvalidArgument, "failed to parse field mask")
	}

	fmt.Println(updateFields)

	updatedVersion, err := s.repo.UpdateVersion(stream.Context(), packageName, versionName, updateFields)
	if err != nil {
		return err
	}

	if err := stream.Send(&polvo_v1.UpdateVersionResponse{
		Version: updatedVersion,
	}); err != nil {
		return err
	}

	return nil
}

func (s *Server) GetPackage(ctx context.Context, request *polvo_v1.GetPackageRequest) (*polvo_v1.GetPackageResponse, error) {
	packageName := parsePackageOrn(request.GetOrn())

	foundPackage, err := s.repo.GetPackage(ctx, packageName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get package")
	}

	return &polvo_v1.GetPackageResponse{
		Package: foundPackage,
	}, nil
}

func (s *Server) GetVersion(ctx context.Context, request *polvo_v1.GetVersionRequest) (*polvo_v1.GetVersionResponse, error) {
	packageName, versionName := parseVersionOrn(request.GetOrn())

	if versionName != defaultVersions["any"] {
		foundPackage, err := s.repo.GetVersion(ctx, packageName, versionName)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get package")
		}

		return &polvo_v1.GetVersionResponse{
			Version: foundPackage,
		}, nil
	}

	foundPackage, err := s.repo.GetHeaviestVersion(ctx, packageName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get package")
	}

	return &polvo_v1.GetVersionResponse{
		Version: foundPackage,
	}, nil
}

func (s *Server) ListPackages(request *polvo_v1.ListPackagesRequest, stream polvo_v1.PolvoService_ListPackagesServer) error {
	packages, err := s.repo.ListPackages(stream.Context())
	if err != nil {
		return errors.Wrap(err, "failed to get package")
	}

	for _, pkg := range packages {
		resp := polvo_v1.ListPackagesResponse{
			Packages: []*polvo_v1.Package{
				pkg,
			},
		}

		if err := stream.Send(&resp); err != nil {
			return errors.Wrap(err, "failed to send to downstream")
		}
	}

	return nil
}

func (s *Server) GetManifestUrl(ctx context.Context, request *polvo_v1.GetManifestUrlRequest) (*polvo_v1.GetManifestUrlResponse, error) {

	packageName, versionName := parseVersionOrn(request.GetOrn())

	version, err := s.repo.GetVersion(ctx, packageName, versionName)
	if err != nil {
		return nil, err
	}

	response := &polvo_v1.GetManifestUrlResponse{
		ManifestUrl: version.GetManifestUrl(),
	}

	return response, nil
}

func (s *Server)ListVersions(request *polvo_v1.ListVersionsRequest, stream polvo_v1.PolvoService_ListVersionsServer) error {
	packageName := parsePackageOrn(request.GetOrn())

	versions, err := s.repo.ListVersions(stream.Context(), packageName)
	if err != nil {
		return errors.Wrap(err, "failed to get package")
	}

	for _, version := range versions {
		resp := polvo_v1.ListVersionsResponse{
			Versions: []*polvo_v1.Version{
				version,
			},
		}

		if err := stream.Send(&resp); err != nil {
			return errors.Wrap(err, "failed to send to downstream")
		}
	}

	return nil
}

func (s *Server) CreateVersion(request *polvo_v1.CreateVersionRequest, stream polvo_v1.PolvoService_CreateVersionServer) error {
	packageName := parsePackageOrn(request.GetPackageOrn())

	isVersionExists, err := s.repo.IsVersionExists(stream.Context(), packageName, request.GetVersion().GetName())
	if err != nil {
		return err
	}

	if isVersionExists {
		return status.Error(codes.NotFound, "Version is exists")
	}

	version := request.GetVersion()

	if version.GetName() == defaultVersions["any"] {
		return status.Errorf(codes.Aborted, "can not create version: any")
	}

	createdVersion, err := s.repo.CreateVersion(stream.Context(), packageName, version)
	if err != nil {
		return status.Errorf(codes.Internal, "failed to create version: %s", err)
	}

	if err := stream.Send(&polvo_v1.CreateVersionResponse{
		Version: createdVersion,
	}); err != nil {
		return status.Errorf(codes.Internal, "failed to send response to client: %s", err)
	}

	return nil
}

func (s *Server)DeleteVersion(request *polvo_v1.DeleteVersionRequest, stream polvo_v1.PolvoService_DeleteVersionServer) error {

	packageName, versionName := parseVersionOrn(request.GetOrn())

	if err := stream.Send(&polvo_v1.DeleteVersionResponse{
		Message: "Version is being detached from package",
	}); err != nil {
		return err
	}

	if err := stream.Send(&polvo_v1.DeleteVersionResponse{
		Message: "Version is detached from package",
	}); err != nil {
		return err
	}

	if err := stream.Send(&polvo_v1.DeleteVersionResponse{
		Message: "Version is being deleted",
	}); err != nil {
		return err
	}

	if err := s.repo.DeleteVersion(stream.Context(), packageName, versionName); err != nil {
		return err
	}

	if err := stream.Send(&polvo_v1.DeleteVersionResponse{
		Message: "Version is deleted",
	}); err != nil {
		return err
	}

	return nil
}
