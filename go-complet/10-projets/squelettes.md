# Squelettes de mini-projets Go

Voici plusieurs squelettes de projets à réaliser pour mettre en pratique vos compétences Go. Chaque squelette contient la structure de base, les fichiers à créer et des idées d’extensions.

---

## 1. Calculatrice CLI

```
calculatrice/
  main.go
```
main.go :
```go
package main
import (
    "fmt"
    "os"
    "strconv"
)
func main() {
    if len(os.Args) < 4 {
        fmt.Println("Usage: calc <nb1> <op> <nb2>")
        return
    }
    a, _ := strconv.Atoi(os.Args[1])
    op := os.Args[2]
    b, _ := strconv.Atoi(os.Args[3])
    var res int
    switch op {
    case "+": res = a + b
    case "-": res = a - b
    case "*": res = a * b
    case "/": res = a / b
    default: fmt.Println("Opérateur inconnu"); return
    }
    fmt.Println("Résultat:", res)
}
```
Extensions : gestion des erreurs, opérations avancées, tests.

---

## 2. Gestionnaire de contacts (JSON)

```
contacts/
  main.go
  contacts.json
```
main.go :
```go
package main
import (
    "encoding/json"
    "fmt"
    "os"
)
type Contact struct {
    Nom  string
    Tel  string
}
func main() {
    // À compléter : menu, ajout, affichage, sauvegarde JSON
}
```
Extensions : recherche, suppression, interface CLI.

---

## 3. Serveur HTTP de gestion de tâches

```
todo-server/
  main.go
```
main.go :
```go
package main
import (
    "encoding/json"
    "net/http"
)
type Task struct {
    ID   int
    Name string
    Done bool
}
var tasks []Task
func main() {
    http.HandleFunc("/tasks", func(w http.ResponseWriter, r *http.Request) {
        json.NewEncoder(w).Encode(tasks)
    })
    http.ListenAndServe(":8080", nil)
}
```
Extensions : ajout/suppression de tâches, persistance, tests, authentification.

---

## 4. Crawler web concurrent

```
crawler/
  main.go
```
main.go :
```go
package main
import (
    "fmt"
    "net/http"
    "sync"
)
func main() {
    urls := []string{"https://golang.org", "https://go.dev"}
    var wg sync.WaitGroup
    for _, url := range urls {
        wg.Add(1)
        go func(u string) {
            defer wg.Done()
            resp, err := http.Get(u)
            if err == nil {
                fmt.Println(u, resp.Status)
            }
        }(url)
    }
    wg.Wait()
}
```
Extensions : exploration récursive, gestion des erreurs, sauvegarde des résultats.

---

## 5. Analyseur de logs système

```
log-analyzer/
  main.go
  logs.txt
```
main.go :
```go
package main
import (
    "bufio"
    "fmt"
    "os"
)
func main() {
    f, _ := os.Open("logs.txt")
    defer f.Close()
    scanner := bufio.NewScanner(f)
    for scanner.Scan() {
        fmt.Println(scanner.Text())
    }
}
```
Extensions : statistiques, filtrage, export CSV.

---

Pour chaque squelette, créez le dossier correspondant, copiez le code de base, puis complétez et améliorez selon les suggestions !
