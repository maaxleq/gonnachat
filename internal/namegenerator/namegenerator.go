package namegenerator

import (
	"github.com/google/uuid"
	"math/rand"
	"strings"
)

const consonants = "bcdfghjklmnpqrstvwxyz"
const vowels = "aeiou"
const nameLength = 8

func Generate() string {
	// Generate a UUID
	genUuid := uuid.New()

	// Strip the hyphens
	strippedUuid := strings.ReplaceAll(genUuid.String(), "-", "")

	// Shuffle the stripped UUID
	uuidRunes := []rune(strippedUuid)
	rand.Shuffle(len(uuidRunes), func(i, j int) {
		uuidRunes[i], uuidRunes[j] = uuidRunes[j], uuidRunes[i]
	})

	shuffledUuid := string(uuidRunes)

	// Generate the name

	name := ""
	for i, runeValue := range shuffledUuid {
		if i >= nameLength {
			break
		}
		if i%2 == 0 {
			// Use consonants
			index := int(runeValue) % len(consonants)
			name += string(consonants[index])
		} else {
			// Use vowels
			index := int(runeValue) % len(vowels)
			name += string(vowels[index])
		}
	}

	return name
}
