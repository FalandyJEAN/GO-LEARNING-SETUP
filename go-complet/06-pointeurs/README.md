# Pointeurs en Go

## Objectifs
- Comprendre et utiliser les pointeurs

## Exemple :
```go
package main
import "fmt"
func incremente(x *int) {
    *x = *x + 1
}
func main() {
    a := 5
    incremente(&a)
    fmt.Println(a) // Affiche 6
}
```

---

Passez à la section suivante une fois à l’aise avec ces notions.