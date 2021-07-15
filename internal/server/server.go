package server

import (
	"context"

	"github.com/golang/protobuf/protoc-gen-go/generator"
	"github.com/google/wire"
	fieldmask_utils "github.com/mennanov/fieldmask-utils"
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

func (s *Server) CreateApplication(request *polvo_v1.CreateApplicationRequest, stream polvo_v1.PolvoService_CreateApplicationServer) error {
	savedApplication, err := s.repo.SetApplication(stream.Context(), request.GetApplication())
	if err != nil {
		return err
	}

	response := &polvo_v1.CreateApplicationResponse{
		Application: savedApplication,
	}

	if err := stream.Send(response); err != nil {
		return err
	}

	return nil
}

func (s *Server) CreatePackage(request *polvo_v1.CreatePackageRequest, stream polvo_v1.PolvoService_CreatePackageServer) error {

	savedPackage, err := s.repo.SetPackage(stream.Context(), request.GetApplicationOrn(), request.GetPackage())
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

func (s *Server) GetPackageEntryPoint(ctx context.Context, request *polvo_v1.GetPackageEntryPointRequest) (*polvo_v1.GetPackageEntryPointResponse, error) {

	versions, err := s.repo.ListVersions(ctx, request.GetPackageOrn())
	if err != nil {
		return nil, errors.Wrap(err, "failed to list version")
	}

	for _, foundVersion := range versions  {

		if foundVersion.BuildName == "latest" {
			return &polvo_v1.GetPackageEntryPointResponse{
				EntryPointUrl: foundVersion.GetEntryPointUrl(),
			}, nil
		}
	}

	return nil, status.Error(codes.NotFound, "no version found")
}

func (s *Server) UpdatePackage(ctx context.Context, request *polvo_v1.UpdatePackageRequest) (*polvo_v1.UpdatePackageResponse, error) {

	request.GetUpdateMask().Normalize()
	if !request.GetUpdateMask().IsValid(request) {
		return nil, errors.New("update mask is invalid")
	}

	filteredRequest := map[string]interface{}{}

	mask, err := fieldmask_utils.MaskFromProtoFieldMask(request.GetUpdateMask(), generator.CamelCase)
	if err != nil {
		return nil, err
	}

	if err := fieldmask_utils.StructToMap(mask, request, filteredRequest); err != nil {
		return nil, err
	}

	updateFields := filteredRequest[string(request.GetPackage().ProtoReflect().Descriptor().Name())].(map[string]interface{})

	savedPackage, err := s.repo.SetPackage(ctx, request.GetPackageOrn(), updateFields)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get package")
	}

	return &polvo_v1.UpdatePackageResponse{
		Package: savedPackage,
	}, nil
}

func (s *Server) GetPackage(ctx context.Context, request *polvo_v1.GetPackageRequest) (*polvo_v1.GetPackageResponse, error) {
	foundPackage, err := s.repo.GetPackage(ctx, request.GetPackageOrn())
	if err != nil {
		return nil, errors.Wrap(err, "failed to get package")
	}

	return &polvo_v1.GetPackageResponse{
		Package: foundPackage,
	}, nil
}

func (s *Server) ListPackages(request *polvo_v1.ListPackagesRequest, streamServer polvo_v1.PolvoService_ListPackagesServer) error {
	packages, err := s.repo.ListPackages(streamServer.Context(), request.GetApplicationOrn())
	if err != nil {
		return errors.Wrap(err, "failed to get package")
	}

	for _, pkg := range packages {
		resp := polvo_v1.ListPackagesResponse{
			Packages: []*polvo_v1.Package{
				pkg,
			},
		}

		if err := streamServer.Send(&resp); err != nil {
			return errors.Wrap(err, "failed to send to downstream")
		}
	}

	return nil
}
