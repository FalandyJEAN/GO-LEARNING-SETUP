# Tableaux, slices et maps

## Objectifs
- Manipuler des tableaux, slices et maps

## Exemple :
```go
package main
import "fmt"
func main() {
    notes := []int{12, 15, 18}
    notes = append(notes, 20)
    for i, note := range notes {
        fmt.Println("Note", i, ":", note)
    }
    capitales := map[string]string{"France": "Paris", "Italie": "Rome"}
    fmt.Println("Capitale de la France:", capitales["France"])
}
```

---

Passez à la section suivante une fois à l’aise avec ces notions.