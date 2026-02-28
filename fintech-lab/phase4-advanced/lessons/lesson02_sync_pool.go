// ================================================================================
// LECON 11 — sync.Pool : Reutilisation d'Objets
// ================================================================================
// COMMENT EXECUTER :  go run phase4-advanced/lessons/lesson02_sync_pool.go
// POUR BENCHMARKER :  go test -bench=. -benchmem phase4-advanced/lessons/
// OBJECTIF         :  Eliminer les allocations repetees en HFT
// ================================================================================

package main

import (
	"fmt"
	"sync"
)

// ============================================================
// PARTIE 1 : PROBLEME — Allocations repetees
// ============================================================
// En HFT, un matching engine peut traiter des MILLIONS de messages
// par seconde. Si on alloue un nouvel objet pour chaque message
// et qu'on le jette ensuite, le GC doit nettoyer des millions
// d'objets -> pauses GC -> latence impredictible.
//
// Solution : sync.Pool

// Message represente un message de marche
type Message struct {
	Type    string
	Symbol  string
	Payload [256]byte // buffer de donnees brutes
}

// ============================================================
// PARTIE 2 : sync.Pool
// ============================================================
// sync.Pool est un cache d'objets reutilisables.
//
// Get()  : recuperer un objet du pool (ou en creer un nouveau)
// Put()  : remettre un objet dans le pool apres utilisation
//
// Le GC PEUT vider le pool entre deux GC cycles.
// -> Le pool n'est PAS un cache permanent, c'est un optimiseur.
//
// ATTENTION : Ne pas stocker d'etat persistant dans un objet pool !
//             Toujours REINITIALISER avant Put() OU apres Get().

var messagePool = sync.Pool{
	// New : fonction appelee quand le pool est vide
	New: func() interface{} {
		return &Message{} // allouer un nouveau Message
	},
}

// ============================================================
// PARTIE 3 : PATTERN D'UTILISATION
// ============================================================

func traiterMessageSansPool(symbol string) {
	// Sans pool : nouvelle allocation a chaque appel
	msg := &Message{} // allocation heap !
	msg.Type = "TICK"
	msg.Symbol = symbol
	// ... traitement ...
	// msg est jete -> GC doit nettoyer
	_ = msg
}

func traiterMessageAvecPool(symbol string) {
	// Avec pool : recuperer un objet existant
	msg := messagePool.Get().(*Message) // type assertion

	// IMPORTANT : toujours reinitialiser avant utilisation !
	// (l'objet peut venir d'un usage precedent)
	msg.Type = "TICK"
	msg.Symbol = symbol
	msg.Payload = [256]byte{} // reinitialiser le buffer

	// ... traitement ...

	// IMPORTANT : remettre dans le pool apres usage
	// Reinitialiser d'abord (eviter les fuites de donnees)
	msg.Type = ""
	msg.Symbol = ""
	messagePool.Put(msg) // retour au pool
}

// ============================================================
// PARTIE 4 : POOL DE BUFFERS (cas le plus courant en HFT)
// ============================================================
// Les buffers d'octets sont les objets les plus souvent poolés.

var bufferPool = sync.Pool{
	New: func() interface{} {
		// Allouer un buffer de 4KB
		buf := make([]byte, 0, 4096)
		return &buf
	},
}

func encoderOrdre(symbol string, prix float64, qty int) []byte {
	// Recuperer un buffer du pool
	bufPtr := bufferPool.Get().(*[]byte)
	buf := (*bufPtr)[:0] // reset la longueur, garde la capacite

	// Ecrire dans le buffer
	buf = fmt.Appendf(buf, "%s,%.4f,%d", symbol, prix, qty)

	// Copier le resultat (le buffer retourne au pool)
	resultat := make([]byte, len(buf))
	copy(resultat, buf)

	// Retourner le buffer au pool
	*bufPtr = buf
	bufferPool.Put(bufPtr)

	return resultat
}

// ============================================================
// PARTIE 5 : POOL TYPEE (pattern Go idiomatique)
// ============================================================
// On cree souvent un type wrapper autour de sync.Pool
// pour avoir un acces type-safe sans type assertion.

type OrdrePool struct {
	pool sync.Pool
}

type Ordre struct {
	ID       int64
	Symbol   string
	Quantity int
	Price    float64
}

func NewOrdrePool() *OrdrePool {
	return &OrdrePool{
		pool: sync.Pool{
			New: func() interface{} { return new(Ordre) },
		},
	}
}

// Get : type-safe, pas besoin de type assertion a l'exterieur
func (p *OrdrePool) Get() *Ordre {
	return p.pool.Get().(*Ordre)
}

// Put : reinitialise et retourne au pool
func (p *OrdrePool) Put(o *Ordre) {
	// Zero-out avant de remettre dans le pool
	*o = Ordre{}
	p.pool.Put(o)
}

// ============================================================
// MAIN
// ============================================================

func main() {

	fmt.Println("=== DEMO sync.Pool ===")

	// Utilisation basique
	for i := 0; i < 5; i++ {
		traiterMessageAvecPool(fmt.Sprintf("AAPL_%d", i))
	}
	fmt.Println("5 messages traites avec pool (zéro allocation supplementaire)")

	// Pool typee
	fmt.Println("\n=== POOL TYPEE (OrdrePool) ===")

	ordrePool := NewOrdrePool()

	// Simuler un traitement haute frequence
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		i := i
		go func() {
			defer wg.Done()

			// Recuperer un ordre du pool
			o := ordrePool.Get()

			// Remplir l'ordre
			o.ID = int64(i)
			o.Symbol = "AAPL"
			o.Quantity = 100 * (i + 1)
			o.Price = 142.50 + float64(i)*0.01

			// Traiter...
			fmt.Printf("  Traitement ordre %d : %s %d@%.2f\n",
				o.ID, o.Symbol, o.Quantity, o.Price)

			// Remettre dans le pool
			ordrePool.Put(o)
		}()
	}
	wg.Wait()

	// Encodage avec buffer pool
	fmt.Println("\n=== POOL DE BUFFERS ===")
	encoded := encoderOrdre("AAPL", 142.5023, 1000)
	fmt.Printf("Encoded : %s\n", encoded)

	// ============================================================
	// RESUME
	// ============================================================
	// 1. sync.Pool evite les allocations repetees en reutilisant des objets
	// 2. pool.Get()  : recuperer (ou creer via New si vide)
	// 3. pool.Put()  : retourner au pool
	// 4. TOUJOURS reinitialiser l'objet avant Put() ou apres Get()
	// 5. Le GC peut vider le pool -> ne pas y stocker d'etat permanent
	// 6. Creer un wrapper typee pour eviter les type assertions
	// 7. Cas typiques : buffers d'octets, messages, ordres, ticks
	// 8. Utiliser avec -benchmem pour mesurer l'impact

	fmt.Println("\n=== FIN LECON 11 ===")
	fmt.Println("Tu peux maintenant attaquer les examens Phase 4 !")
	fmt.Println("  exam_escape_analysis.go")
	fmt.Println("  exam_sync_pool.go")
}
