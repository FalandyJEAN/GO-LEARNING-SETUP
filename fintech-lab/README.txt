================================================================================
  FINTECH LAB — Go HFT Learning Laboratory
  Role: Staff Engineer / Technical Interviewer
  Department: High Frequency Trading & Low Latency Systems
================================================================================

OBJECTIF
--------
Passer de zero a un niveau technique suffisant pour reussir des entretiens
de type "Quantitative Developer" ou "Low-Latency Go Engineer" dans une
banque d'investissement (tier-1: JPMorgan, Goldman Sachs, Citadel).

PHILOSOPHIE
-----------
- Pas de QCM. Uniquement du code reel a deboguer, corriger, optimiser.
- Chaque phase debloque la suivante. Pas de raccourci.
- Le code review est intraitable. Un commentaire vague = rejet.

STRUCTURE DU LABORATOIRE
-------------------------

fintech-lab/
|
|-- README.txt                          <- Ce fichier
|
|-- phase1-bootcamp/                    <- Zero to 1 : Syntaxe et outils
|   |-- bootcamp.txt                    <- Plan d'action et commandes setup
|   |-- main.go                         <- Ton premier script de validation
|   |-- exams/
|       |-- exam01_syntax.go            <- EXAMEN 1 : Code casse (syntax/logic)
|       |-- exam01_instructions.txt     <- Instructions de l'examen 1
|
|-- phase2-order-engine/                <- Projet fil rouge : Moteur de matching
|   |-- specs/
|       |-- architecture.txt            <- Architecture du systeme
|       |-- data_structures.txt         <- Structures de donnees financieres
|
|-- phase3-concurrency/                 <- Goroutines, Channels, Mutex, Races
|   |-- exams/
|       |-- exam_data_race.go           <- Code avec data race silencieuse
|       |-- exam_memory_leak.go         <- Code avec fuite memoire
|
|-- phase4-advanced/                    <- GC tuning, sync.Pool, escape analysis
|   |-- notes_gc.txt                    <- Notes sur l'optimisation du GC

REGLES D'ENGAGEMENT
--------------------
1. Tu soumets ton code corrige dans la conversation.
2. Je fais la code review. Si c'est insuffisant, je te renvoie une v2.
3. Chaque concept valide = passage a la phase suivante.
4. Les delais d'entretien ne pardonnent pas. La vitesse compte.

PREREQUIS
----------
- Go 1.22+ installe : https://go.dev/dl/
- Verifier : go version
- Editeur recommande : VSCode + extension "Go" (golang.go)
- Terminal : bash ou PowerShell

COMMANDES D'INITIALISATION (a executer une seule fois)
-------------------------------------------------------
  cd fintech-lab
  go mod init fintech-lab
  go run phase1-bootcamp/main.go

================================================================================
