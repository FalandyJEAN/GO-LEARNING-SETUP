# Les bases du langage Go

## Objectifs
- Comprendre les variables, types, opérateurs
- Utiliser les conditions et les boucles

## Sommaire
- Variables et constantes
- Types de base (int, float, string, bool)
- Opérateurs
- Structures de contrôle (if, switch)
- Boucles (for)

## Exemple :
```go
package main
import "fmt"
func main() {
    var age int = 30
    nom := "Alice"
    if age > 18 {
        fmt.Println(nom, "est majeur(e)")
    }
    for i := 0; i < 3; i++ {
        fmt.Println("Compteur:", i)
    }
}
```

---

Passez à la section suivante une fois à l’aise avec ces notions.