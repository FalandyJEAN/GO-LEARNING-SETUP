# 05. Tableaux, Slices et Maps en Go

## Objectifs pédagogiques
- Manipuler efficacement les collections de données
- Comprendre les différences entre tableaux, slices et maps
- Utiliser les fonctions de la librairie standard pour traiter les collections

---

## 1. Théorie et concepts
- **Tableau** : taille fixe, typé
- **Slice** : tableau dynamique, flexible
- **Map** : dictionnaire clé/valeur

## 2. Exemples de code
### Slice dynamique
```go
nombres := []int{1, 2, 3}
nombres = append(nombres, 4)
```

### Map clé/valeur
```go
capitales := map[string]string{"France": "Paris", "Italie": "Rome"}
```

## 3. Exercices pratiques
- Créez une slice de notes et calculez la moyenne
- Utilisez une map pour compter les occurrences de mots dans une phrase
- Triez une slice d’entiers

## 4. Mini-projets
- Gestionnaire de contacts (ajout, recherche, suppression)
- Analyseur de texte (statistiques sur les mots)

## 5. Quiz & approfondissement
1. Quelle est la différence entre un tableau et une slice ?
2. Comment supprimer un élément d’une map ?
3. Peut-on avoir une slice de slices ?

## 6. Ressources complémentaires
- [Go by Example : Slices](https://gobyexample.com/slices)
- [Go by Example : Maps](https://gobyexample.com/maps)
- [Effective Go : Slices](https://go.dev/doc/effective_go#slices)

---

Passez à la section suivante après avoir réalisé un mini-projet.