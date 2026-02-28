# 04. Structures, Méthodes et Interfaces en Go

## Objectifs pédagogiques
- Comprendre la programmation orientée objet façon Go (structs, méthodes, interfaces)
- Utiliser la composition et l’encapsulation
- Appliquer les patterns de conception Go

---

## 1. Théorie et concepts
- **Struct** : structure de données personnalisée
- **Méthode** : fonction attachée à un type
- **Interface** : contrat de comportement
- **Composition** : inclusion de structs dans d’autres structs

## 2. Exemples de code
### Définir une struct et une méthode
```go
type Voiture struct {
    Marque string
    Vitesse int
}
func (v *Voiture) Accelerer() {
    v.Vitesse += 10
}
```

### Interface et polymorphisme
```go
type Parleur interface {
    Parler() string
}
type Humain struct{}
func (h Humain) Parler() string { return "Bonjour !" }
```

## 3. Exercices pratiques
- Créez une struct Animal avec une méthode Parler
- Implémentez une interface pour différents types d’animaux
- Utilisez la composition pour créer une struct Véhicule avec un moteur

## 4. Mini-projets
- Système de gestion de bibliothèque (livres, utilisateurs, emprunts)
- Simulateur de zoo (animaux, enclos, interactions)

## 5. Quiz & approfondissement
1. Quelle est la différence entre une méthode et une fonction ?
2. Comment Go gère-t-il l’héritage ?
3. À quoi sert une interface vide ?

## 6. Ressources complémentaires
- [Go by Example : Structs](https://gobyexample.com/structs)
- [Go by Example : Interfaces](https://gobyexample.com/interfaces)
- [Effective Go : Embedding](https://go.dev/doc/effective_go#embedding)

---

Passez à la section suivante après avoir réalisé un mini-projet.