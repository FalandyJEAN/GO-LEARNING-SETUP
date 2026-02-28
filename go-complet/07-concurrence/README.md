# Concurrence en Go

## Objectifs
- Utiliser les goroutines
- Communiquer avec les channels
- Synchroniser avec les mutex

## Exemple :
```go
package main
import (
    "fmt"
    "time"
)
func travail(nom string) {
    for i := 0; i < 3; i++ {
        fmt.Println(nom, "tour", i)
        time.Sleep(100 * time.Millisecond)
    }
}
func main() {
    go travail("A")
    travail("B")
    time.Sleep(500 * time.Millisecond)
}
```

---

Passez à la section suivante une fois à l’aise avec ces notions.