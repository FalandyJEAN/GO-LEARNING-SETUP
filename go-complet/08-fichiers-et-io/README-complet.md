# 08. Fichiers et Entrées/Sorties en Go

## Objectifs pédagogiques
- Lire et écrire dans des fichiers
- Manipuler les flux d’entrée/sortie (stdin, stdout)
- Sérialiser/désérialiser des données (JSON, CSV)

---

## 1. Théorie et concepts
- **os.File** : manipulation de fichiers
- **bufio** : lecture/écriture efficace
- **encoding/json/csv** : sérialisation

## 2. Exemples de code
### Écriture dans un fichier
```go
f, _ := os.Create("test.txt")
f.WriteString("Bonjour fichier!")
f.Close()
```

### Lecture d’un fichier
```go
data, _ := os.ReadFile("test.txt")
fmt.Println(string(data))
```

### Sérialisation JSON
```go
type Personne struct { Nom string }
p := Personne{"Alice"}
b, _ := json.Marshal(p)
```

## 3. Exercices pratiques
- Lire un fichier texte et compter les lignes
- Écrire un programme qui copie un fichier
- Sérialiser une struct en JSON et la sauvegarder

## 4. Mini-projets
- Carnet d’adresses sauvegardé en JSON
- Analyseur de logs

## 5. Quiz & approfondissement
1. Quelle est la différence entre os.Open et os.Create ?
2. Comment lire un fichier ligne par ligne ?
3. Pourquoi utiliser bufio ?

## 6. Ressources complémentaires
- [Go by Example : Files](https://gobyexample.com/files)
- [Go by Example : JSON](https://gobyexample.com/json)
- [Effective Go : I/O](https://go.dev/doc/effective_go#io)

---

Passez à la section suivante après avoir réalisé un mini-projet.