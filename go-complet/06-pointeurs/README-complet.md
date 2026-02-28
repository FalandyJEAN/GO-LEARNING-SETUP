# 06. Pointeurs et Références en Go

## Objectifs pédagogiques
- Comprendre la gestion mémoire en Go
- Utiliser les pointeurs pour optimiser les performances
- Éviter les pièges courants liés aux références

---

## 1. Théorie et concepts
- **Pointeur** : variable contenant l’adresse mémoire d’une autre variable
- **Référence** : passage par adresse
- **Garbage Collector** : gestion automatique de la mémoire

## 2. Exemples de code
### Utilisation d’un pointeur
```go
a := 5
pa := &a
*pa = 10
```

### Passage par référence
```go
func incremente(x *int) { *x++ }
```

## 3. Exercices pratiques
- Écrivez une fonction qui échange deux variables via des pointeurs
- Créez une struct Personne et modifiez-la via un pointeur

## 4. Mini-projets
- Gestionnaire de mémoire simulé (allocation/libération)
- Manipulation de matrices via pointeurs

## 5. Quiz & approfondissement
1. Quelle est la différence entre un pointeur et une valeur ?
2. Pourquoi Go n’a-t-il pas d’arithmétique de pointeur ?
3. Quand utiliser un pointeur sur une struct ?

## 6. Ressources complémentaires
- [Go by Example : Pointers](https://gobyexample.com/pointers)
- [Effective Go : Pointers](https://go.dev/doc/effective_go#pointers)

---

Passez à la section suivante après avoir réalisé un mini-projet.