# Fonctions et gestion des erreurs

## Objectifs
- Définir et utiliser des fonctions
- Comprendre la portée des variables
- Gérer les erreurs

## Exemple :
```go
package main
import (
    "fmt"
    "errors"
)
func division(a, b float64) (float64, error) {
    if b == 0 {
        return 0, errors.New("division par zéro")
    }
    return a / b, nil
}
func main() {
    res, err := division(10, 0)
    if err != nil {
        fmt.Println("Erreur:", err)
    } else {
        fmt.Println("Résultat:", res)
    }
}
```

---

Passez à la section suivante une fois à l’aise avec ces notions.