package api

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
	db "github.com/mitch-jensen/mymovies/dbstore"
)

const (
	movieReleasesPath = "/movies/{movieId}/releases"
	releaseByIDPath   = "/releases/{id}"
)

func (s *Server) registerReleaseRoutes() {
	huma.Register(s.api, huma.Operation{
		OperationID: "list-movie-releases",
		Method:      http.MethodGet,
		Path:        movieReleasesPath,
		Summary:     "List the home video releases of a movie",
	}, s.listMovieReleases)

	huma.Register(s.api, huma.Operation{
		OperationID:   "create-release",
		Method:        http.MethodPost,
		Path:          movieReleasesPath,
		Summary:       "Add a home video release to a movie",
		DefaultStatus: http.StatusCreated,
	}, s.createRelease)

	huma.Register(s.api, huma.Operation{
		OperationID: "get-release",
		Method:      http.MethodGet,
		Path:        releaseByIDPath,
		Summary:     "Get a home video release by ID",
	}, s.getRelease)

	huma.Register(s.api, huma.Operation{
		OperationID: "update-release",
		Method:      http.MethodPut,
		Path:        releaseByIDPath,
		Summary:     "Update a home video release",
	}, s.updateRelease)

	huma.Register(s.api, huma.Operation{
		OperationID:   "delete-release",
		Method:        http.MethodDelete,
		Path:          releaseByIDPath,
		Summary:       "Delete a home video release",
		DefaultStatus: http.StatusNoContent,
	}, s.deleteRelease)
}

// Release is the API representation of a physical home video release of a movie.
type Release struct {
	ID           uuid.UUID        `json:"id"`
	MovieID      uuid.UUID        `json:"movieId"`
	Studio       *string          `json:"studio,omitempty"`
	CountryCode  *string          `json:"countryCode,omitempty"`
	UPC          *string          `json:"upc,omitempty"`
	EAN          *string          `json:"ean,omitempty"`
	ASIN         *string          `json:"asin,omitempty"`
	ReleaseDate  *time.Time       `json:"releaseDate,omitempty"`
	Casing       *string          `json:"casing,omitempty"`
	Slipcover    *bool            `json:"slipcover,omitempty"`
	BluRayDiscs  *int32           `json:"bluRayDiscs,omitempty"`
	DVDDiscs     *int32           `json:"dvdDiscs,omitempty"`
	DigitalCopy  bool             `json:"digitalCopy"`
	CreatedAt    time.Time        `json:"createdAt"`
	Watched      bool             `json:"watched"`
	Comment      *string          `json:"comment,omitempty"`
	Retailer     *string          `json:"retailer,omitempty"`
	Price        *decimal.Decimal `json:"price,omitempty"`
	PriceComment *string          `json:"priceComment,omitempty"`
}

func releaseFromDB(release db.HomeVideoRelease) Release {
	return Release{
		ID:           release.ID,
		MovieID:      release.MovieID,
		Studio:       release.Studio,
		CountryCode:  release.CountryCode,
		UPC:          release.Upc,
		EAN:          release.Ean,
		ASIN:         release.Asin,
		ReleaseDate:  release.ReleaseDate,
		Casing:       release.Casing,
		Slipcover:    release.Slipcover,
		BluRayDiscs:  release.BluRayDiscs,
		DVDDiscs:     release.DvdDiscs,
		DigitalCopy:  release.DigitalCopy,
		CreatedAt:    release.CreatedAt,
		Watched:      release.Watched,
		Comment:      release.Comment,
		Retailer:     release.Retailer,
		Price:        release.Price,
		PriceComment: release.PriceComment,
	}
}

// ReleaseFields holds the editable fields of a home video release. The owning
// movie is fixed at creation and identified by the URL path, not the body.
type ReleaseFields struct {
	Studio       *string          `json:"studio,omitempty"`
	CountryCode  *string          `json:"countryCode,omitempty"`
	UPC          *string          `json:"upc,omitempty"`
	EAN          *string          `json:"ean,omitempty"`
	ASIN         *string          `json:"asin,omitempty"`
	ReleaseDate  *time.Time       `json:"releaseDate,omitempty"`
	Casing       *string          `json:"casing,omitempty"`
	Slipcover    *bool            `json:"slipcover,omitempty"`
	BluRayDiscs  *int32           `json:"bluRayDiscs,omitempty"`
	DVDDiscs     *int32           `json:"dvdDiscs,omitempty"`
	DigitalCopy  bool             `json:"digitalCopy"`
	Watched      bool             `json:"watched"`
	Comment      *string          `json:"comment,omitempty"`
	Retailer     *string          `json:"retailer,omitempty"`
	Price        *decimal.Decimal `json:"price,omitempty"`
	PriceComment *string          `json:"priceComment,omitempty"`
}

// ReleaseIDInput identifies a release by its path parameter.
type ReleaseIDInput struct {
	ID uuid.UUID `doc:"Release ID" path:"id"`
}

// MovieReleasesInput identifies a movie whose releases are addressed.
type MovieReleasesInput struct {
	MovieID uuid.UUID `doc:"Movie ID" path:"movieId"`
}

// CreateReleaseInput carries the owning movie ID in the path and the release
// fields in the body.
type CreateReleaseInput struct {
	MovieID uuid.UUID `doc:"Movie ID" path:"movieId"`
	Body    ReleaseFields
}

// UpdateReleaseInput carries the release ID in the path and new field values in
// the body.
type UpdateReleaseInput struct {
	ID   uuid.UUID `doc:"Release ID" path:"id"`
	Body ReleaseFields
}

// ReleaseOutput is a response carrying a single release.
type ReleaseOutput struct {
	Body Release
}

// ListReleasesOutput is the response body for listing a movie's releases.
type ListReleasesOutput struct {
	Body []Release
}

func (s *Server) listMovieReleases(ctx context.Context, input *MovieReleasesInput) (*ListReleasesOutput, error) {
	releases, err := s.queries.ListHomeVideoReleasesByMovie(ctx, input.MovieID)
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to list releases", err)
	}

	body := make([]Release, len(releases))
	for i, release := range releases {
		body[i] = releaseFromDB(release)
	}

	return &ListReleasesOutput{Body: body}, nil
}

func (s *Server) createRelease(ctx context.Context, input *CreateReleaseInput) (*ReleaseOutput, error) {
	// Confirm the movie exists so a missing movie yields 404 rather than a raw
	// foreign-key error.
	_, err := s.queries.GetMovie(ctx, input.MovieID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, huma.Error404NotFound("movie not found")
		}

		return nil, huma.Error500InternalServerError("failed to look up movie", err)
	}

	release, err := s.queries.CreateHomeVideoRelease(ctx, createParams(input.MovieID, input.Body))
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to create release", err)
	}

	return &ReleaseOutput{Body: releaseFromDB(release)}, nil
}

func (s *Server) getRelease(ctx context.Context, input *ReleaseIDInput) (*ReleaseOutput, error) {
	release, err := s.queries.GetHomeVideoRelease(ctx, input.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, huma.Error404NotFound("release not found")
		}

		return nil, huma.Error500InternalServerError("failed to get release", err)
	}

	return &ReleaseOutput{Body: releaseFromDB(release)}, nil
}

func (s *Server) updateRelease(ctx context.Context, input *UpdateReleaseInput) (*ReleaseOutput, error) {
	release, err := s.queries.UpdateHomeVideoRelease(ctx, updateParams(input.ID, input.Body))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, huma.Error404NotFound("release not found")
		}

		return nil, huma.Error500InternalServerError("failed to update release", err)
	}

	return &ReleaseOutput{Body: releaseFromDB(release)}, nil
}

func (s *Server) deleteRelease(ctx context.Context, input *ReleaseIDInput) (*struct{}, error) {
	err := s.queries.DeleteHomeVideoRelease(ctx, input.ID)
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to delete release", err)
	}

	return nil, nil //nolint:nilnil // 204 No Content: no body and no error.
}

func createParams(movieID uuid.UUID, fields ReleaseFields) db.CreateHomeVideoReleaseParams {
	return db.CreateHomeVideoReleaseParams{
		MovieID:      movieID,
		Studio:       fields.Studio,
		CountryCode:  fields.CountryCode,
		Upc:          fields.UPC,
		Ean:          fields.EAN,
		Asin:         fields.ASIN,
		ReleaseDate:  fields.ReleaseDate,
		Casing:       fields.Casing,
		Slipcover:    fields.Slipcover,
		BluRayDiscs:  fields.BluRayDiscs,
		DvdDiscs:     fields.DVDDiscs,
		DigitalCopy:  fields.DigitalCopy,
		Watched:      fields.Watched,
		Comment:      fields.Comment,
		Retailer:     fields.Retailer,
		Price:        fields.Price,
		PriceComment: fields.PriceComment,
	}
}

func updateParams(id uuid.UUID, fields ReleaseFields) db.UpdateHomeVideoReleaseParams {
	return db.UpdateHomeVideoReleaseParams{
		ID:           id,
		Studio:       fields.Studio,
		CountryCode:  fields.CountryCode,
		Upc:          fields.UPC,
		Ean:          fields.EAN,
		Asin:         fields.ASIN,
		ReleaseDate:  fields.ReleaseDate,
		Casing:       fields.Casing,
		Slipcover:    fields.Slipcover,
		BluRayDiscs:  fields.BluRayDiscs,
		DvdDiscs:     fields.DVDDiscs,
		DigitalCopy:  fields.DigitalCopy,
		Watched:      fields.Watched,
		Comment:      fields.Comment,
		Retailer:     fields.Retailer,
		Price:        fields.Price,
		PriceComment: fields.PriceComment,
	}
}
