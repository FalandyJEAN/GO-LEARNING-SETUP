# 07. Concurrence et Programmation Parallèle en Go

> Maîtrisez la programmation concurrente, un des points forts de Go, pour écrire des applications performantes, robustes et modernes.

## Objectifs pédagogiques
- Comprendre la différence entre concurrence et parallélisme
- Utiliser les goroutines pour exécuter du code en parallèle
- Communiquer entre routines avec les channels
- Synchroniser l’accès aux ressources partagées (mutex, sync)
- Appliquer les paradigmes de la programmation concurrente (producteur/consommateur, fan-in/fan-out, worker pool)
- Gérer les erreurs et les contextes d’annulation

---

## 1. Théorie et concepts
- **Concurrence vs Parallélisme** : Go permet d’exécuter plusieurs tâches quasi-simultanément (concurrence), qui peuvent être réparties sur plusieurs cœurs (parallélisme).
- **Goroutine** : Légère unité d’exécution gérée par le runtime Go.
- **Channel** : Outil de communication sécurisé entre goroutines.
- **Mutex & sync** : Synchronisation d’accès aux ressources partagées.
- **Context** : Gestion de l’annulation, des délais et de la propagation d’informations entre goroutines.

## 2. Exemples de code
### Goroutine simple
```go
go fmt.Println("Bonjour depuis une goroutine !")
```

### Channel de communication
```go
ch := make(chan int)
go func() { ch <- 42 }()
val := <-ch
fmt.Println(val) // Affiche 42
```

### Mutex
```go
var mu sync.Mutex
mu.Lock()
// section critique
mu.Unlock()
```

### Pattern : Worker Pool
```go
// ...exemple complet dans worker_pool.go...
```

## 3. Exercices pratiques
- Créez un programme qui lance 10 goroutines, chacune affiche son numéro.
- Implémentez un compteur partagé entre plusieurs goroutines (avec et sans mutex).
- Réalisez un pipeline de traitement avec plusieurs étapes connectées par des channels.

## 4. Mini-projets
- **Serveur de chat concurrent** : chaque client est géré par une goroutine, messages transmis via channels.
- **Crawler web parallèle** : exploration de liens en parallèle, synchronisation des résultats.
- **Simulation de files d’attente (banque, supermarché)** : chaque client/goroutine, gestion des ressources partagées.

## 5. Quiz & approfondissement
1. Quelle est la différence entre un thread OS et une goroutine ?
2. Que se passe-t-il si on lit/écrit sur un channel sans buffer sans lecteur/écrivain ?
3. Pourquoi utiliser un mutex ?
4. Comment annuler proprement une goroutine ?

## 6. Ressources complémentaires
- [Tour of Go : Concurrency](https://tour.golang.org/concurrency/1)
- [Go by Example : Concurrency](https://gobyexample.com/concurrency)
- [Effective Go : Concurrency](https://go.dev/doc/effective_go#concurrency)

---

Passez à la section suivante une fois à l’aise avec ces notions et après avoir réalisé au moins un mini-projet.
