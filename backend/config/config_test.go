package config

import (
	"net/url"
	"strings"
	"testing"
)

func TestLoadMongoURIEscapesCredentials(t *testing.T) {
	t.Setenv("MONGODB_URI", "")
	t.Setenv("MONGO_USER", "cinema@example.com")
	t.Setenv("MONGO_PASSWORD", "p@ss:/word")
	t.Setenv("MONGO_HOST", "mongodb:27017")
	t.Setenv("MONGODB_DB", "cinemadb")
	t.Setenv("MONGO_REPLICA_SET", "rs-dev")
	t.Setenv("MONGO_DIRECT_CONNECTION", "true")

	uri := loadMongoURI()
	if strings.Contains(uri, "cinema@example.com") || strings.Contains(uri, "p@ss:/word") {
		t.Fatalf("credentials were not escaped: %s", uri)
	}
	parsed, err := url.Parse(uri)
	if err != nil {
		t.Fatalf("parse MongoDB URI: %v", err)
	}
	if parsed.Host != "mongodb:27017" || parsed.Path != "/cinemadb" {
		t.Fatalf("unexpected MongoDB address: %s", uri)
	}
	if parsed.Query().Get("authSource") != "admin" ||
		parsed.Query().Get("replicaSet") != "rs-dev" ||
		parsed.Query().Get("directConnection") != "true" {
		t.Fatalf("unexpected MongoDB options: %s", uri)
	}
}

func TestLoadMongoURIAddsOptionsForLocalExplicitURI(t *testing.T) {
	t.Setenv("MONGODB_URI", "mongodb://user:password@localhost:27017/cinemadb?authSource=admin")
	t.Setenv("MONGO_REPLICA_SET", "")
	t.Setenv("MONGO_DIRECT_CONNECTION", "")

	parsed, err := url.Parse(loadMongoURI())
	if err != nil {
		t.Fatalf("parse MongoDB URI: %v", err)
	}
	if parsed.Query().Get("replicaSet") != "rs0" ||
		parsed.Query().Get("directConnection") != "true" {
		t.Fatalf("local MongoDB options were not added: %s", parsed.Redacted())
	}
}

func TestLoadMongoURIDoesNotModifyRemoteExplicitURI(t *testing.T) {
	const uri = "mongodb+srv://user:password@example.mongodb.net/cinemadb?retryWrites=true"
	t.Setenv("MONGODB_URI", uri)

	if got := loadMongoURI(); got != uri {
		t.Fatalf("remote MongoDB URI was modified: %s", got)
	}
}
