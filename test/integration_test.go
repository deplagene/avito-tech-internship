package test

import (
	"bytes"
	"context"
	"deplagene/avito-tech-internship/cmd/api"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"testing"
	"time"
)

const (
	appURL = "http://localhost:8080"
	dbURL  = "postgres://user:password@localhost:5432/avito_trainee_test?sslmode=disable"
)

func TestMain(m *testing.M) {
	// Устанавливаем переменную окружения для тестовой БД
	os.Setenv("DATABASE_URL", dbURL)
	os.Setenv("ENV", "test")
	os.Setenv("POSTGRES_DB", "avito_trainee_test")

	// Запускаем docker-compose
	fmt.Println("Starting docker-compose for integration tests...")
	cmd := exec.Command("docker-compose", "up", "--build", "-d")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("Failed to start docker-compose: %v\n", err)
		os.Exit(1)
	}

	// Ожидаем готовность приложения
	fmt.Println("Waiting for application to be ready...")
	if err := waitForApp(appURL+"/health", 60*time.Second); err != nil {
		fmt.Printf("Application not ready: %v\n", err)
		teardownTestEnvironment()
		os.Exit(1)
	}
	fmt.Println("Application is ready.")

	// Запускаем тесты
	code := m.Run()

	// Останавливаем docker-compose
	fmt.Println("Tearing down test environment...")
	teardownTestEnvironment()

	os.Exit(code)
}

func waitForApp(url string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	client := &http.Client{}
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for app at %s", url)
		default:
			req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
			resp, err := client.Do(req)
			if err == nil && resp.StatusCode == http.StatusOK {
				resp.Body.Close()
				return nil
			}
			if resp != nil {
				resp.Body.Close()
			}
			time.Sleep(1 * time.Second)
		}
	}
}

func teardownTestEnvironment() {
	cmd := exec.Command("docker-compose", "down", "-v")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error tearing down docker-compose: %v\n", err)
	}
}

func clearDatabase() {
	// TODO: Реализовать очистку базы данных перед каждым тестом
}

func makeRequest(t *testing.T, method, path string, body any) (*http.Response, []byte) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("Failed to marshal request body: %v", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, appURL+path, reqBody)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	return resp, respBytes
}

func TestPostTeamAdd(t *testing.T) {
	// ! Пока заглушка
	clearDatabase()

	teamName := "test-team"
	user1ID := "u1"
	user2ID := "u2"

	teamData := api.Team{
		TeamName: teamName,
		Members: []api.TeamMember{
			{UserId: user1ID, Username: "Alice", IsActive: true},
			{UserId: user2ID, Username: "Bob", IsActive: true},
		},
	}

	resp, respBody := makeRequest(t, "POST", "/team/add", teamData)

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusCreated, resp.StatusCode, respBody)
	}

	var response struct {
		Team api.Team `json:"team"`
	}
	if err := json.Unmarshal(respBody, &response); err != nil {
		t.Fatalf("Failed to unmarshal response body: %v", err)
	}

	if response.Team.TeamName != teamName {
		t.Errorf("Expected team name %s, got %s", teamName, response.Team.TeamName)
	}
	if len(response.Team.Members) != 2 {
		t.Errorf("Expected 2 members, got %d", len(response.Team.Members))
	}
}

// TODO: Добавить другие интеграционные тесты для всех эндпоинтов
// TestGetTeamGet
// TestPostUsersSetIsActive
// TestPostPullRequestCreate
// TestPostPullRequestMerge
// TestPostPullRequestReassign
// TestGetUsersGetReview
