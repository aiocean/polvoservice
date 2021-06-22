package server

import (
	"context"
	"strings"

	"cloud.google.com/go/firestore"
	"github.com/golang/protobuf/protoc-gen-go/generator"
	"github.com/google/wire"
	fieldmask_utils "github.com/mennanov/fieldmask-utils"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	polvo_v1 "pkg.aiocean.dev/polvogo/aiocean/polvo/v1"
	"pkg.aiocean.dev/polvoservice/internal/repository"
	v1 "pkg.aiocean.dev/polvoservice/pkg/client/v1"
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

func (s *Server) CreatePackage(request *polvo_v1.CreatePackageRequest, stream polvo_v1.PolvoService_CreatePackageServer) error {

	if err := s.repo.SetResource(stream.Context(), request.GetPackage().GetOrn(), request.GetPackage()); err != nil {
		return errors.Wrap(err, "failed to get package")
	}

	savedPackage, err := s.repo.GetPackage(stream.Context(), request.GetPackage().GetOrn())
	if err != nil {
		return err
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

	versionsRef, err := s.repo.ListVersions(ctx, request.GetPackageOrn())
	if err != nil {
		return nil, errors.Wrap(err, "failed to list version")
	}

	for {
		snapshot, err := versionsRef.Next()
		if err == iterator.Done {
			break
		}

		if err != nil {
			return nil, errors.Wrap(err, "failed to get nex")
		}

		var foundVersion polvo_v1.Version
		if err := snapshot.DataTo(&foundVersion); err != nil {
			return nil, errors.Wrap(err, "failed to decode data")
		}

		foundVersion.Orn = v1.ServiceDomain + "/" + strings.Join(strings.Split(snapshot.Ref.Path, "/")[5:], "/")

		if strings.HasSuffix(foundVersion.Orn, "/latest") {
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

	if err := s.repo.SetResource(ctx, request.GetPackage().GetOrn(), updateFields, firestore.MergeAll); err != nil {
		return nil, errors.Wrap(err, "failed to get package")
	}

	savedPackage, err := s.repo.GetPackage(ctx, request.GetPackage().GetOrn())
	if err != nil {
		return nil, err
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
	packagesRef, err := s.repo.ListPackages(streamServer.Context(), request.GetApplicationOrn())
	if err != nil {
		return errors.Wrap(err, "failed to get package")
	}

	for {
		snapshot, err := packagesRef.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return errors.Wrap(err, "failed to get nex")
		}

		var foundPackage polvo_v1.Package
		if err := snapshot.DataTo(&foundPackage); err != nil {
			return errors.Wrap(err, "failed to decode data")
		}

		foundPackage.Orn = v1.ServiceDomain + "/" + strings.Join(strings.Split(snapshot.Ref.Path, "/")[5:], "/")

		resp := polvo_v1.ListPackagesResponse{
			Packages: []*polvo_v1.Package{
				&foundPackage,
			},
		}

		if err := streamServer.Send(&resp); err != nil {
			return errors.Wrap(err, "failed to send to downstream")
		}
	}

	return nil
}

func (s *Server) CreateVersion(ctx context.Context, request *polvo_v1.CreateVersionRequest) (*polvo_v1.CreateVersionResponse, error) {

	packageOrn := strings.Split(request.GetVersion().GetOrn(), "/versions/")[0]

	if _, err := s.repo.GetPackage(ctx, packageOrn); err != nil {
		return nil, errors.Wrap(err, "failed to get package")
	}

	if err := s.repo.SetResource(ctx, request.GetVersion().GetOrn(), request.GetVersion()); err != nil {
		return nil, errors.Wrap(err, "failed to get version")
	}

	return &polvo_v1.CreateVersionResponse{
		Version: request.Version,
	}, nil
}

func (s *Server) DeleteVersion(request *polvo_v1.DeleteVersionRequest, stream polvo_v1.PolvoService_DeleteVersionServer) error {
	response := &polvo_v1.DeleteVersionResponse{Message: ""}

	response.Message = "Start delete version"
	if err := stream.Send(response); err != nil {
		return err
	}

	if err := s.repo.DeleteResource(stream.Context(), request.GetVersionOrn()); err != nil {
		return err
	}

	response.Message = "Version delete success"
	if err := stream.Send(response); err != nil {
		return err
	}

	return nil
}

func (s *Server) DeletePackage(request *polvo_v1.DeletePackageRequest, streamServer polvo_v1.PolvoService_DeletePackageServer) error {
	response := &polvo_v1.DeletePackageResponse{Message: ""}

	response.Message = "Delete versions success"
	if err := streamServer.Send(response); err != nil {
		return err
	}

	response.Message = "Delete build success"
	if err := streamServer.Send(response); err != nil {
		return err
	}

	if err := s.repo.DeleteResource(streamServer.Context(), request.GetPackageOrn()); err != nil {
		return status.Error(codes.Internal, "failed to delete package: "+request.GetPackageOrn()+", err: "+err.Error())
	}

	response.Message = "Delete package success"
	if err := streamServer.Send(response); err != nil {
		return err
	}

	return nil
}

func (s *Server) UpdateVersion(ctx context.Context, request *polvo_v1.UpdateVersionRequest) (*polvo_v1.UpdateVersionResponse, error) {

	packageOrn := strings.Split(request.GetVersion().GetOrn(), "/versions/")[0]

	if _, err := s.repo.GetPackage(ctx, packageOrn); err != nil {
		return nil, errors.Wrap(err, "failed to get package")
	}

	request.GetUpdateMask().Normalize()
	if !request.GetUpdateMask().IsValid(request) {
		return nil, errors.New("update mask is invalid")
	}

	filter, err := fieldmask_utils.MaskFromProtoFieldMask(request.GetUpdateMask(), generator.CamelCase)
	if err != nil {
		return nil, err
	}

	filteredRequest := map[string]interface{}{}

	if err := fieldmask_utils.StructToMap(filter, request, filteredRequest); err != nil {
		return nil, err
	}

	updateFields := filteredRequest[string(request.GetVersion().ProtoReflect().Descriptor().Name())].(map[string]interface{})

	if err := s.repo.SetResource(ctx, request.GetVersion().GetOrn(), updateFields, firestore.MergeAll); err != nil {
		return nil, errors.Wrap(err, "failed to get version")
	}

	savedVersion, err := s.repo.GetVersion(ctx, request.GetVersion().GetOrn())
	if err != nil {
		return nil, err
	}

	return &polvo_v1.UpdateVersionResponse{
		Version: savedVersion,
	}, nil
}

func (s *Server) GetVersion(ctx context.Context, request *polvo_v1.GetVersionRequest) (*polvo_v1.GetVersionResponse, error) {
	foundVersion, err := s.repo.GetVersion(ctx, request.GetVersionOrn())
	if err != nil {
		return nil, err
	}

	return &polvo_v1.GetVersionResponse{
		Version: foundVersion,
	}, nil
}

func (s *Server) ListVersions(request *polvo_v1.ListVersionsRequest, stream polvo_v1.PolvoService_ListVersionsServer) error {
	versionsRef, err := s.repo.ListVersions(stream.Context(), request.GetPackageOrn())
	if err != nil {
		return errors.Wrap(err, "failed to list version")
	}

	for {
		snapshot, err := versionsRef.Next()
		if err == iterator.Done {
			break
		}

		if err != nil {
			return errors.Wrap(err, "failed to get nex")
		}

		var foundVersion polvo_v1.Version
		if err := snapshot.DataTo(&foundVersion); err != nil {
			return errors.Wrap(err, "failed to decode data")
		}

		foundVersion.Orn = v1.ServiceDomain + "/" + strings.Join(strings.Split(snapshot.Ref.Path, "/")[5:], "/")
		resp := polvo_v1.ListVersionsResponse{
			Versions: []*polvo_v1.Version{
				&foundVersion,
			},
		}

		if err := stream.Send(&resp); err != nil {
			return errors.Wrap(err, "failed to send to downstream")
		}
	}

	return nil
}
