package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/media-parser/backend/internal/config"
	"github.com/media-parser/backend/internal/model/entity"
	"github.com/media-parser/backend/internal/repository/postgres"
)

func main() {
	name := flag.String("name", "", "Имя ключа (обязательно)")
	expires := flag.String("expires", "", "Срок действия (например: 24h, 7d, 30d, never)")
	allowParse := flag.Bool("parse", true, "Право на парсинг")
	allowMedia := flag.Bool("media", true, "Право на чтение медиа")
	allowRequests := flag.Bool("requests", true, "Право на просмотр запросов")
	flag.Parse()

	if *name == "" {
		fmt.Println("Использование:")
		fmt.Println("  go run cmd/keygen/main.go -name <имя> [опции]")
		fmt.Println("")
		fmt.Println("Опции:")
		fmt.Println("  -name      Имя ключа (обязательно)")
		fmt.Println("  -expires   Срок действия: 24h, 7d, 30d, never (по умолчанию: never)")
		fmt.Println("  -parse     Право на парсинг (по умолчанию: true)")
		fmt.Println("  -media     Право на чтение медиа (по умолчанию: true)")
		fmt.Println("  -requests  Право на просмотр запросов (по умолчанию: true)")
		fmt.Println("")
		fmt.Println("Примеры:")
		fmt.Println("  go run cmd/keygen/main.go -name \"frontend-app\"")
		fmt.Println("  go run cmd/keygen/main.go -name \"temp-key\" -expires 24h")
		fmt.Println("  go run cmd/keygen/main.go -name \"readonly\" -parse=false")
		os.Exit(1)
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := postgres.NewPostgresDB(&cfg.Database)
	if err != nil && cfg.Database.Host == "postgres" {
		log.Printf("Failed to connect to %s, trying localhost...", cfg.Database.Host)
		cfg.Database.Host = "localhost"
		db, err = postgres.NewPostgresDB(&cfg.Database)
	}
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	tokenRepo := postgres.NewAPITokenRepository(db)

	key := generateSecureKey()
	tokenName := fmt.Sprintf("%s-%s", *name, uuid.New().String()[:8])

	permissions := entity.TokenPermissions{
		Parse:     *allowParse,
		MediaRead: *allowMedia,
		RequestsView: *allowRequests,
	}

	permData, err := json.Marshal(permissions)
	if err != nil {
		log.Fatalf("Failed to marshal permissions: %v", err)
	}

	var expiresAt *time.Time
	if *expires != "" && *expires != "never" {
		duration, err := parseDuration(*expires)
		if err != nil {
			log.Fatalf("Invalid expires value: %v", err)
		}
		t := time.Now().Add(duration)
		expiresAt = &t
	}

	token := &entity.APIToken{
		Token:       key,
		Name:        &tokenName,
		Active:      true,
		ExpiresAt:   expiresAt,
		Permissions: permData,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := tokenRepo.Create(ctx, token); err != nil {
		log.Fatalf("Failed to create API token: %v", err)
	}

	fmt.Println("============================================")
	fmt.Println("API ключ успешно создан!")
	fmt.Println("============================================")
	fmt.Printf("ID:          %d\n", token.ID)
	fmt.Printf("Имя:         %s\n", tokenName)
	fmt.Printf("Ключ:        %s\n", key)
	fmt.Printf("Создан:      %s\n", time.Now().Format(time.RFC3339))
	if expiresAt != nil {
		fmt.Printf("Истекает:    %s\n", expiresAt.Format(time.RFC3339))
	} else {
		fmt.Printf("Истекает:    никогда\n")
	}
	fmt.Println("")
	fmt.Println("Права доступа:")
	fmt.Printf("  Parse:    %v\n", *allowParse)
	fmt.Printf("  Media:    %v\n", *allowMedia)
	fmt.Printf("  Requests: %v\n", *allowRequests)
	fmt.Println("============================================")
	fmt.Println("")
	fmt.Println("Используйте этот ключ в заголовке X-Auth-Token")
}

func generateSecureKey() string {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		log.Fatalf("Failed to generate random bytes: %v", err)
	}
	return hex.EncodeToString(bytes)
}

func parseDuration(s string) (time.Duration, error) {
	s = strings.TrimSpace(s)
	
	if strings.HasSuffix(s, "d") {
		days, err := strconv.Atoi(strings.TrimSuffix(s, "d"))
		if err != nil {
			return 0, err
		}
		return time.Duration(days) * 24 * time.Hour, nil
	}
	
	if strings.HasSuffix(s, "h") {
		hours, err := strconv.Atoi(strings.TrimSuffix(s, "h"))
		if err != nil {
			return 0, err
		}
		return time.Duration(hours) * time.Hour, nil
	}
	
	return time.ParseDuration(s)
}
