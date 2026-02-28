# Tests unitaires en Go

## Objectifs
- Écrire des tests unitaires
- Lancer les tests avec `go test`

## Exemple :
```go
package maths
func Addition(a, b int) int {
    return a + b
}
```

Fichier de test :
```go
package maths
testimport "testing"
func TestAddition(t *testing.T) {
    res := Addition(2, 3)
    if res != 5 {
        t.Errorf("Attendu 5, obtenu %d", res)
    }
}
```

---

Passez à la section suivante une fois à l’aise avec ces notions.