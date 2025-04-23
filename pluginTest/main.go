package main

import (
	"fmt"
	"log"
	"plugin"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwt"
)

func main() {
	// Pfad zur .so Plugin-Datei
	plug, err := plugin.Open("../.engines/.vault/crypto-provider-hashicorp-vault-plugin.so")
	if err != nil {
		log.Fatalf("❌ Plugin konnte nicht geladen werden: %v", err)
	}

	// Exportierte Funktion "Hello" laden
	symHello, err := plug.Lookup("Hello")
	if err != nil {
		log.Fatalf("❌ Symbol 'Hello' nicht gefunden: %v", err)
	}

	// Funktion casten
	helloFunc, ok := symHello.(func() string)
	if !ok {
		log.Fatalf("❌ Symbol hat falschen Typ (erwartet: func() string)")
	}

	// Funktion aufrufen
	result := helloFunc()
	fmt.Println("👋 Plugin sagt:", result)

	_, err = jwt.NewBuilder().
		Issuer("plugin.example").
		Subject("test").
		Expiration(time.Now().Add(1 * time.Hour)).
		Build()
	if err != nil {
		log.Printf("🚨 JWT-Erstellung fehlgeschlagen: %s", err)

	}
}
