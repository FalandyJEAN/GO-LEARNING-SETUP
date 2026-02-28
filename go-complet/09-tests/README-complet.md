# 09. Tests, Qualité et Bonnes Pratiques en Go

## Objectifs pédagogiques
- Écrire des tests unitaires et d’intégration
- Utiliser les outils de test et de benchmark
- Organiser son code et documenter

---

## 1. Théorie et concepts
- **testing** : package standard pour les tests
- **Benchmark** : mesure de performance
- **Organisation** : structure des dossiers, modules

## 2. Exemples de code
### Test unitaire
```go
func Addition(a, b int) int { return a + b }
// Fichier addition_test.go
func TestAddition(t *testing.T) {
    if Addition(2, 3) != 5 {
        t.Error("Erreur de calcul")
    }
}
```

### Benchmark
```go
func BenchmarkAddition(b *testing.B) {
    for i := 0; i < b.N; i++ {
        Addition(2, 3)
    }
}
```

## 3. Exercices pratiques
- Écrire des tests pour une fonction de tri
- Mesurer la performance d’une fonction
- Organiser un projet Go modulaire

## 4. Mini-projets
- Calculatrice testée et documentée
- API REST avec tests d’intégration

## 5. Quiz & approfondissement
1. Quelle est la différence entre un test unitaire et un test d’intégration ?
2. Comment lancer tous les tests d’un projet ?
3. Pourquoi documenter son code ?

## 6. Ressources complémentaires
- [Go by Example : Testing](https://gobyexample.com/testing)
- [Effective Go : Testing](https://go.dev/doc/effective_go#testing)
- [GoDoc](https://pkg.go.dev/golang.org/x/tools/cmd/godoc)

---

Passez à la section suivante après avoir réalisé un mini-projet.