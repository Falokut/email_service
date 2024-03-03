package screeningsservice

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	_ "time/tzdata"

	cinema_service "github.com/Falokut/cinema_service/pkg/cinema_service/v1/protos"
	"github.com/Falokut/email_service/internal/config"
	"github.com/Falokut/email_service/internal/models"
	movies_service "github.com/Falokut/movies_service/pkg/movies_service/v1/protos"
	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	"github.com/opentracing/opentracing-go"
	"github.com/ringsaturn/tzf"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

type ScreeningsService struct {
	cinemaServiceConn   *grpc.ClientConn
	cinemaServiceClient cinema_service.CinemaServiceV1Client
	moviesServiceConn   *grpc.ClientConn
	moviesServiceClient movies_service.MoviesServiceV1Client
	logger              *logrus.Logger
	tzfinder            tzf.F
}

func getGrpcConnection(addr string, cfg config.ConnectionSecureConfig) (*grpc.ClientConn, error) {
	creds, err := cfg.GetGrpcTransportCredentials()
	if err != nil {
		return nil, err
	}

	return grpc.Dial(addr, creds,
		grpc.WithUnaryInterceptor(
			otgrpc.OpenTracingClientInterceptor(opentracing.GlobalTracer())),
		grpc.WithStreamInterceptor(
			otgrpc.OpenTracingStreamClientInterceptor(opentracing.GlobalTracer())),
	)
}

func NewScreeningsService(cinemaServiceAddr string,
	cinemaServiceSecureConfig config.ConnectionSecureConfig,
	moviesServiceAddr string,
	moviesServiceSecureConfig config.ConnectionSecureConfig,
	logger *logrus.Logger) (*ScreeningsService, error) {

	tzfinder, err := tzf.NewDefaultFinder()
	if err != nil {
		return nil, err
	}
	cinemaServiceConn, err := getGrpcConnection(cinemaServiceAddr, cinemaServiceSecureConfig)
	if err != nil {
		return nil, err
	}

	moviesServiceConn, err := getGrpcConnection(moviesServiceAddr, moviesServiceSecureConfig)
	if err != nil {
		cinemaServiceConn.Close()
		return nil, err
	}

	return &ScreeningsService{
		cinemaServiceClient: cinema_service.NewCinemaServiceV1Client(cinemaServiceConn),
		cinemaServiceConn:   cinemaServiceConn,
		moviesServiceClient: movies_service.NewMoviesServiceV1Client(moviesServiceConn),
		moviesServiceConn:   moviesServiceConn,
		logger:              logger,
		tzfinder:            tzfinder,
	}, nil
}

func (s *ScreeningsService) Shutdown() {
	if s.cinemaServiceConn != nil {
		err := s.cinemaServiceConn.Close()
		if err != nil {
			s.logger.Error("error while closing cinema service connection ", err)
		}
	}
	if s.moviesServiceConn != nil {
		err := s.moviesServiceConn.Close()
		if err != nil {
			s.logger.Error("error while closing movies service connection ", err)
		}
	}
}

func (s *ScreeningsService) GetScreeningInfo(ctx context.Context, screeningId int64) (screening models.Screening, err error) {
	defer s.handleError(ctx, &err, "GetScreeningInfo")

	mask := &fieldmaskpb.FieldMask{}
	mask.Paths = []string{"cinema_id", "movie_id", "screening_type", "hall_id", "start_time"}

	res, err := s.cinemaServiceClient.GetScreening(ctx, &cinema_service.GetScreeningRequest{
		ScreeningId: screeningId,
		Mask:        mask})

	if err != nil {
		return
	}

	errCh := make(chan error, 1)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		info, err := s.getCinemaInfo(ctx, res.CinemaId)
		if err != nil {
			errCh <- err
			return
		}

		screening.Cinema = info
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		name, err := s.getHallName(ctx, res.HallId)
		if err != nil {
			errCh <- err
			return
		}

		screening.HallName = name
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		name, posterUrl, err := s.getMovieInfo(ctx, res.MovieId)
		if err != nil {
			errCh <- err
			return
		}

		screening.MovieName = name
		screening.MoviePosterUrl = posterUrl
	}()

	go func() {
		wg.Wait()
		close(errCh)
	}()

LOOP:
	for {
		select {
		case <-ctx.Done():
			return
		case chVal, ok := <-errCh:
			if !ok {
				err = nil
				break LOOP
			}
			if chVal != nil {
				err = chVal
				return
			}
		}
	}

	timezone := s.tzfinder.GetTimezoneName(screening.Cinema.Coordinates.Long, screening.Cinema.Coordinates.Lat)
	tz, err := time.LoadLocation(timezone)
	if err != nil {
		return
	}

	startTime, _ := time.Parse(time.RFC3339, res.StartTime.FormattedTimestamp)
	startTime = startTime.In(tz)

	screening.StartTime = startTime.Format("15:04")
	screening.StartDate = startTime.Format("02.01")

	return
}

func (s *ScreeningsService) getCinemaInfo(ctx context.Context, cinemaId int32) (info models.Cinema, err error) {
	defer s.handleError(ctx, &err, "getCinemaInfo")

	res, err := s.cinemaServiceClient.GetCinema(ctx, &cinema_service.GetCinemaRequest{
		CinemaId: cinemaId,
	})
	if err != nil {
		return
	}

	info = models.Cinema{
		Address: res.Address,
		Name:    res.Name,
		Coordinates: models.Coordinates{
			Long: res.Coordinates.Longitude,
			Lat:  res.Coordinates.Latityde,
		},
	}
	return
}

func (s *ScreeningsService) getHallName(ctx context.Context, hallId int32) (name string, err error) {
	defer s.handleError(ctx, &err, "getHallName")

	res, err := s.cinemaServiceClient.GetHalls(ctx,
		&cinema_service.GetHallsRequest{HallsIds: fmt.Sprint(hallId)})
	if err != nil {
		return
	}
	if len(res.Halls) != 1 {
		err = models.Error(models.NotFound, "hall not found")
		return
	}

	return res.Halls[0].Name, nil
}

func (s *ScreeningsService) getMovieInfo(ctx context.Context, movieId int32) (name string, posterUrl string, err error) {
	defer s.handleError(ctx, &err, "getMovieInfo")
	mask := &fieldmaskpb.FieldMask{}
	mask.Paths = []string{"title_ru", "poster_url"}

	res, err := s.moviesServiceClient.GetMovie(ctx, &movies_service.GetMovieRequest{
		MovieID: movieId,
		Mask:    mask,
	})
	if err != nil {
		return
	}

	return res.TitleRu, res.PosterUrl, nil
}

func (s *ScreeningsService) logError(err error, functionName string) {
	if err == nil {
		return
	}

	var sericeErr = &models.ServiceError{}
	if errors.As(err, &sericeErr) {
		s.logger.WithFields(
			logrus.Fields{
				"error.function.name": functionName,
				"error.msg":           sericeErr.Msg,
				"code":                sericeErr.Code,
			},
		).Error("screenings service error occurred")
	} else {
		s.logger.WithFields(
			logrus.Fields{
				"error.function.name": functionName,
				"error.msg":           err.Error(),
			},
		).Error("screenings service error occurred")
	}

}

func (s *ScreeningsService) handleError(ctx context.Context, err *error, functionName string) {
	if ctx.Err() != nil {
		var code models.ErrorCode
		switch {
		case errors.Is(ctx.Err(), context.Canceled):
			code = models.Canceled
		case errors.Is(ctx.Err(), context.DeadlineExceeded):
			code = models.DeadlineExceeded
		}
		*err = models.Error(code, ctx.Err().Error())
		return
	}
	if err == nil || *err == nil {
		return
	}

	e := *err
	s.logError(*err, functionName)
	switch status.Code(*err) {
	case codes.Canceled:
		*err = models.Error(models.Canceled, e.Error())
	case codes.DeadlineExceeded:
		*err = models.Error(models.DeadlineExceeded, e.Error())
	case codes.Internal:
		*err = models.Error(models.Internal, "")
	case codes.NotFound:
		*err = models.Error(models.NotFound, "screening with specified id not found")
	default:
		*err = models.Error(models.Unknown, e.Error())
	}
}
