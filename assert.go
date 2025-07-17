package fsrs

import "fmt"

func assertCard(card *Card) {
	if !(card.Stability == 0 && card.Difficulty == 0 || card.Difficulty != 0 && card.Stability != 0) {
		panic(fmt.Sprintf("The Difficulty and Stability of a card are either both zero or both non-zero."))

	}
}
