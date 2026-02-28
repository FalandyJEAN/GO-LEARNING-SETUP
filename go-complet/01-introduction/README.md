# Introduction à Go

Bienvenue dans la première leçon !

## Objectifs
- Comprendre l'origine et la philosophie de Go
- Installer Go sur votre machine
- Écrire et exécuter votre premier programme Go

## Ressources
- [Site officiel de Go](https://golang.org)
- [Documentation Go](https://golang.org/doc/)

## À faire
1. Installer Go (voir https://golang.org/doc/install)
2. Vérifier l'installation avec la commande :
   ```sh
   go version
   ```
3. Créer un fichier `hello.go` avec le code suivant :
   ```go
   package main
   import "fmt"
   func main() {
       fmt.Println("Bonjour, Go !")
   }
   ```
4. Exécuter le programme :
   ```sh
   go run hello.go
   ```

---

Passez à la section suivante une fois ces étapes terminées.