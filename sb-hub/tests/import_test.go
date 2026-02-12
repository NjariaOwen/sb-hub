package tests

import (
	"testing"

	"github.com/NjariaOwen/sb-hub/cmd"
	"gopkg.in/yaml.v3"
)

func TestComposeProject_ParseYAML(t *testing.T) {
	input := `
services:
  web:
    image: nginx:latest
    ports:
      - "8080:80"
    volumes:
      - ./data:/var/www
    environment:
      APP_ENV: production
  db:
    image: postgres:15
    ports:
      - "5432:5432"
    environment:
      POSTGRES_PASSWORD: secret
`

	var project cmd.ComposeProject
	err := yaml.Unmarshal([]byte(input), &project)
	if err != nil {
		t.Fatalf("failed to parse YAML: %v", err)
	}

	if len(project.Services) != 2 {
		t.Fatalf("expected 2 services, got %d", len(project.Services))
	}

	web, ok := project.Services["web"]
	if !ok {
		t.Fatal("expected 'web' service")
	}
	if web.Image != "nginx:latest" {
		t.Fatalf("expected web image 'nginx:latest', got '%s'", web.Image)
	}
	if len(web.Ports) != 1 || web.Ports[0] != "8080:80" {
		t.Fatalf("expected web ports ['8080:80'], got %v", web.Ports)
	}
	if len(web.Volumes) != 1 || web.Volumes[0] != "./data:/var/www" {
		t.Fatalf("expected web volumes ['./data:/var/www'], got %v", web.Volumes)
	}
	if web.Env["APP_ENV"] != "production" {
		t.Fatalf("expected APP_ENV=production, got '%s'", web.Env["APP_ENV"])
	}

	db, ok := project.Services["db"]
	if !ok {
		t.Fatal("expected 'db' service")
	}
	if db.Image != "postgres:15" {
		t.Fatalf("expected db image 'postgres:15', got '%s'", db.Image)
	}
	if db.Env["POSTGRES_PASSWORD"] != "secret" {
		t.Fatalf("expected POSTGRES_PASSWORD=secret, got '%s'", db.Env["POSTGRES_PASSWORD"])
	}
}

func TestComposeProject_EmptyServices(t *testing.T) {
	input := `
services: {}
`

	var project cmd.ComposeProject
	err := yaml.Unmarshal([]byte(input), &project)
	if err != nil {
		t.Fatalf("failed to parse YAML: %v", err)
	}
	if len(project.Services) != 0 {
		t.Fatalf("expected 0 services, got %d", len(project.Services))
	}
}

func TestComposeProject_InvalidYAML(t *testing.T) {
	input := `
services:
  web:
    image: nginx
    ports: [[[invalid
`

	var project cmd.ComposeProject
	err := yaml.Unmarshal([]byte(input), &project)
	if err == nil {
		t.Fatal("expected error for invalid YAML, got nil")
	}
}

func TestComposeProject_SingleService(t *testing.T) {
	input := `
services:
  app:
    image: node:18
    ports:
      - "3000:3000"
`

	var project cmd.ComposeProject
	err := yaml.Unmarshal([]byte(input), &project)
	if err != nil {
		t.Fatalf("failed to parse YAML: %v", err)
	}

	if len(project.Services) != 1 {
		t.Fatalf("expected 1 service, got %d", len(project.Services))
	}
	app := project.Services["app"]
	if app.Image != "node:18" {
		t.Fatalf("expected image 'node:18', got '%s'", app.Image)
	}
}

func TestComposeProject_NoImage(t *testing.T) {
	input := `
services:
  worker:
    ports:
      - "8080:80"
`

	var project cmd.ComposeProject
	err := yaml.Unmarshal([]byte(input), &project)
	if err != nil {
		t.Fatalf("failed to parse YAML: %v", err)
	}

	worker := project.Services["worker"]
	if worker.Image != "" {
		t.Fatalf("expected empty image, got '%s'", worker.Image)
	}
}

func TestComposeProject_MultipleVolumes(t *testing.T) {
	input := `
services:
  app:
    image: alpine
    volumes:
      - ./code:/app
      - ./data:/data
      - ./config:/etc/app
`

	var project cmd.ComposeProject
	err := yaml.Unmarshal([]byte(input), &project)
	if err != nil {
		t.Fatalf("failed to parse YAML: %v", err)
	}

	app := project.Services["app"]
	if len(app.Volumes) != 3 {
		t.Fatalf("expected 3 volumes, got %d", len(app.Volumes))
	}
}
