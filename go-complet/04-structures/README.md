# Structures, méthodes et interfaces

## Objectifs
- Définir des structs
- Attacher des méthodes
- Utiliser les interfaces

## Exemple :
```go
package main
import "fmt"
type Animal struct {
    Nom string
}
func (a Animal) Parler() {
    fmt.Println("Je suis", a.Nom)
}
func main() {
    chat := Animal{Nom: "Minou"}
    chat.Parler()
}
```

---

Passez à la section suivante une fois à l’aise avec ces notions.