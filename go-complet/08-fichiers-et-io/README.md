# Fichiers et entrées/sorties

## Objectifs
- Lire et écrire dans des fichiers
- Utiliser l’entrée/sortie standard

## Exemple :
```go
package main
import (
    "fmt"
    "os"
)
func main() {
    f, err := os.Create("test.txt")
    if err != nil {
        fmt.Println(err)
        return
    }
    defer f.Close()
    f.WriteString("Bonjour fichier!")
}
```

---

Passez à la section suivante une fois à l’aise avec ces notions.