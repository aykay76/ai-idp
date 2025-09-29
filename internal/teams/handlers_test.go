package teams

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/aykay76/ai-idp/internal/logger"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockTeamService is a mock implementation of the team service for testing
type MockTeamService struct {
	mock.Mock
}

func (m *MockTeamService) CreateTeam(ctx context.Context, team Team) (Team, error) {
	args := m.Called(ctx, team)
	return args.Get(0).(Team), args.Error(1)
}

func (m *MockTeamService) GetTeam(ctx context.Context, teamID uuid.UUID) (Team, error) {
	args := m.Called(ctx, teamID)
	return args.Get(0).(Team), args.Error(1)
}

func (m *MockTeamService) ListTeams(ctx context.Context, limit, offset int) ([]Team, int, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]Team), args.Int(1), args.Error(2)
}

func (m *MockTeamService) UpdateTeam(ctx context.Context, team Team) (Team, error) {
	args := m.Called(ctx, team)
	return args.Get(0).(Team), args.Error(1)
}

func (m *MockTeamService) DeleteTeam(ctx context.Context, teamID uuid.UUID) error {
	args := m.Called(ctx, teamID)
	return args.Error(0)
}

func setupTestHandlers() (*Handlers, *MockTeamService) {
	mockService := &MockTeamService{}
	testLogger := logger.New("debug", "text")
	handlers := NewHandlers(mockService, testLogger)
	return handlers, mockService
}

func TestHandlers_CreateTeam(t *testing.T) {
	handlers, mockService := setupTestHandlers()

	t.Run("successful creation", func(t *testing.T) {
		team := Team{
			Name:        "test-team",
			DisplayName: "Test Team",
			Description: stringPtr("Test description"),
			LeadEmail:   "lead@company.com",
		}

		expectedTeam := team
		expectedTeam.ID = uuid.New()
		expectedTeam.CreatedAt = time.Now().UTC()
		expectedTeam.UpdatedAt = expectedTeam.CreatedAt

		mockService.On("CreateTeam", mock.Anything, mock.AnythingOfType("Team")).Return(expectedTeam, nil).Once()

		reqBody, err := json.Marshal(team)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/teams", bytes.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		
		rr := httptest.NewRecorder()
		handlers.CreateTeam(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

		var responseTeam Team
		err = json.Unmarshal(rr.Body.Bytes(), &responseTeam)
		require.NoError(t, err)
		assert.Equal(t, expectedTeam.ID, responseTeam.ID)
		assert.Equal(t, expectedTeam.Name, responseTeam.Name)

		mockService.AssertExpectations(t)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/teams", strings.NewReader("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		
		rr := httptest.NewRecorder()
		handlers.CreateTeam(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)

		var errorResp ErrorResponse
		err := json.Unmarshal(rr.Body.Bytes(), &errorResp)
		require.NoError(t, err)
		assert.Equal(t, "Invalid JSON in request body", errorResp.Message)
		assert.Equal(t, "INVALID_JSON", errorResp.Code)
	})

	t.Run("service error", func(t *testing.T) {
		team := Team{
			Name:        "test-team",
			DisplayName: "Test Team",
			LeadEmail:   "lead@company.com",
		}

		mockService.On("CreateTeam", mock.Anything, mock.AnythingOfType("Team")).Return(Team{}, ErrInvalidTeamData).Once()

		reqBody, err := json.Marshal(team)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/teams", bytes.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		
		rr := httptest.NewRecorder()
		handlers.CreateTeam(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)

		var errorResp ErrorResponse
		err = json.Unmarshal(rr.Body.Bytes(), &errorResp)
		require.NoError(t, err)
		assert.Equal(t, "Failed to create team", errorResp.Message)
		assert.Equal(t, "CREATE_FAILED", errorResp.Code)

		mockService.AssertExpectations(t)
	})
}

func TestHandlers_GetTeam(t *testing.T) {
	handlers, mockService := setupTestHandlers()

	t.Run("successful get", func(t *testing.T) {
		teamID := uuid.New()
		expectedTeam := Team{
			ID:          teamID,
			Name:        "test-team",
			DisplayName: "Test Team",
			LeadEmail:   "lead@company.com",
			CreatedAt:   time.Now().UTC(),
		}

		mockService.On("GetTeam", mock.Anything, teamID).Return(expectedTeam, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/"+teamID.String(), nil)
		req.SetPathValue("id", teamID.String())
		
		rr := httptest.NewRecorder()
		handlers.GetTeam(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

		var responseTeam Team
		err := json.Unmarshal(rr.Body.Bytes(), &responseTeam)
		require.NoError(t, err)
		assert.Equal(t, expectedTeam.ID, responseTeam.ID)
		assert.Equal(t, expectedTeam.Name, responseTeam.Name)

		mockService.AssertExpectations(t)
	})

	t.Run("missing team ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/", nil)
		
		rr := httptest.NewRecorder()
		handlers.GetTeam(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)

		var errorResp ErrorResponse
		err := json.Unmarshal(rr.Body.Bytes(), &errorResp)
		require.NoError(t, err)
		assert.Equal(t, "Team ID is required", errorResp.Message)
		assert.Equal(t, "MISSING_TEAM_ID", errorResp.Code)
	})

	t.Run("invalid team ID format", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/invalid-uuid", nil)
		req.SetPathValue("id", "invalid-uuid")
		
		rr := httptest.NewRecorder()
		handlers.GetTeam(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)

		var errorResp ErrorResponse
		err := json.Unmarshal(rr.Body.Bytes(), &errorResp)
		require.NoError(t, err)
		assert.Equal(t, "Invalid team ID format", errorResp.Message)
		assert.Equal(t, "INVALID_TEAM_ID", errorResp.Code)
	})

	t.Run("team not found", func(t *testing.T) {
		teamID := uuid.New()
		mockService.On("GetTeam", mock.Anything, teamID).Return(Team{}, ErrTeamNotFound).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/"+teamID.String(), nil)
		req.SetPathValue("id", teamID.String())
		
		rr := httptest.NewRecorder()
		handlers.GetTeam(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)

		var errorResp ErrorResponse
		err := json.Unmarshal(rr.Body.Bytes(), &errorResp)
		require.NoError(t, err)
		assert.Equal(t, "Team not found", errorResp.Message)
		assert.Equal(t, "TEAM_NOT_FOUND", errorResp.Code)

		mockService.AssertExpectations(t)
	})
}

func TestHandlers_ListTeams(t *testing.T) {
	handlers, mockService := setupTestHandlers()

	t.Run("successful list", func(t *testing.T) {
		expectedTeams := []Team{
			{
				ID:          uuid.New(),
				Name:        "team1",
				DisplayName: "Team 1",
				LeadEmail:   "lead1@company.com",
			},
			{
				ID:          uuid.New(),
				Name:        "team2",
				DisplayName: "Team 2",
				LeadEmail:   "lead2@company.com",
			},
		}

		mockService.On("ListTeams", mock.Anything, 50, 0).Return(expectedTeams, 2, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/v1/teams", nil)
		
		rr := httptest.NewRecorder()
		handlers.ListTeams(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

		var response ListTeamsResponse
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Len(t, response.Teams, 2)
		assert.Equal(t, 50, response.Pagination.Limit)
		assert.Equal(t, 0, response.Pagination.Offset)
		assert.Equal(t, 2, response.Pagination.Total)

		mockService.AssertExpectations(t)
	})

	t.Run("with pagination parameters", func(t *testing.T) {
		expectedTeams := []Team{
			{
				ID:          uuid.New(),
				Name:        "team3",
				DisplayName: "Team 3",
				LeadEmail:   "lead3@company.com",
			},
		}

		mockService.On("ListTeams", mock.Anything, 10, 20).Return(expectedTeams, 25, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/v1/teams?limit=10&offset=20", nil)
		
		rr := httptest.NewRecorder()
		handlers.ListTeams(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var response ListTeamsResponse
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Len(t, response.Teams, 1)
		assert.Equal(t, 10, response.Pagination.Limit)
		assert.Equal(t, 20, response.Pagination.Offset)
		assert.Equal(t, 25, response.Pagination.Total)

		mockService.AssertExpectations(t)
	})

	t.Run("service error", func(t *testing.T) {
		mockService.On("ListTeams", mock.Anything, 50, 0).Return([]Team{}, 0, assert.AnError).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/v1/teams", nil)
		
		rr := httptest.NewRecorder()
		handlers.ListTeams(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)

		var errorResp ErrorResponse
		err := json.Unmarshal(rr.Body.Bytes(), &errorResp)
		require.NoError(t, err)
		assert.Equal(t, "Failed to list teams", errorResp.Message)
		assert.Equal(t, "LIST_FAILED", errorResp.Code)

		mockService.AssertExpectations(t)
	})
}

func TestHandlers_UpdateTeam(t *testing.T) {
	handlers, mockService := setupTestHandlers()

	t.Run("successful update", func(t *testing.T) {
		teamID := uuid.New()
		team := Team{
			Name:        "updated-team",
			DisplayName: "Updated Team",
			LeadEmail:   "updated-lead@company.com",
		}

		expectedTeam := team
		expectedTeam.ID = teamID
		expectedTeam.UpdatedAt = time.Now().UTC()

		mockService.On("UpdateTeam", mock.Anything, mock.MatchedBy(func(t Team) bool {
			return t.ID == teamID
		})).Return(expectedTeam, nil).Once()

		reqBody, err := json.Marshal(team)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPut, "/api/v1/teams/"+teamID.String(), bytes.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.SetPathValue("id", teamID.String())
		
		rr := httptest.NewRecorder()
		handlers.UpdateTeam(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

		var responseTeam Team
		err = json.Unmarshal(rr.Body.Bytes(), &responseTeam)
		require.NoError(t, err)
		assert.Equal(t, expectedTeam.ID, responseTeam.ID)
		assert.Equal(t, expectedTeam.Name, responseTeam.Name)

		mockService.AssertExpectations(t)
	})

	t.Run("team not found", func(t *testing.T) {
		teamID := uuid.New()
		team := Team{
			Name:        "not-found-team",
			DisplayName: "Not Found Team",
			LeadEmail:   "notfound@company.com",
		}

		mockService.On("UpdateTeam", mock.Anything, mock.MatchedBy(func(t Team) bool {
			return t.ID == teamID
		})).Return(Team{}, ErrTeamNotFound).Once()

		reqBody, err := json.Marshal(team)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPut, "/api/v1/teams/"+teamID.String(), bytes.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.SetPathValue("id", teamID.String())
		
		rr := httptest.NewRecorder()
		handlers.UpdateTeam(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)

		var errorResp ErrorResponse
		err = json.Unmarshal(rr.Body.Bytes(), &errorResp)
		require.NoError(t, err)
		assert.Equal(t, "Team not found", errorResp.Message)
		assert.Equal(t, "TEAM_NOT_FOUND", errorResp.Code)

		mockService.AssertExpectations(t)
	})
}

func TestHandlers_DeleteTeam(t *testing.T) {
	handlers, mockService := setupTestHandlers()

	t.Run("successful deletion", func(t *testing.T) {
		teamID := uuid.New()
		mockService.On("DeleteTeam", mock.Anything, teamID).Return(nil).Once()

		req := httptest.NewRequest(http.MethodDelete, "/api/v1/teams/"+teamID.String(), nil)
		req.SetPathValue("id", teamID.String())
		
		rr := httptest.NewRecorder()
		handlers.DeleteTeam(rr, req)

		assert.Equal(t, http.StatusNoContent, rr.Code)
		assert.Empty(t, rr.Body.String())

		mockService.AssertExpectations(t)
	})

	t.Run("team not found", func(t *testing.T) {
		teamID := uuid.New()
		mockService.On("DeleteTeam", mock.Anything, teamID).Return(ErrTeamNotFound).Once()

		req := httptest.NewRequest(http.MethodDelete, "/api/v1/teams/"+teamID.String(), nil)
		req.SetPathValue("id", teamID.String())
		
		rr := httptest.NewRecorder()
		handlers.DeleteTeam(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)

		var errorResp ErrorResponse
		err := json.Unmarshal(rr.Body.Bytes(), &errorResp)
		require.NoError(t, err)
		assert.Equal(t, "Team not found", errorResp.Message)
		assert.Equal(t, "TEAM_NOT_FOUND", errorResp.Code)

		mockService.AssertExpectations(t)
	})
}

// Compile-time check that MockTeamService implements TeamService
var _ TeamService = (*MockTeamService)(nil)