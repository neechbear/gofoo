package main

import (
	"errors"
	"fmt"
)

type SkinColour struct {
	Freckles bool
	Colour   string
}

type Boobies struct {
	Size    int
	Nipples int
	Shade   SkinColour
}

func (b Boobies) squeeze(duration int) (didItLactate bool, err error) {
	if duration <= 0 {
		err = errors.New("You cannot squeeze boobies for a zero or negative period of time")
	}
	return true, err
}

func main() {
	b := new(Boobies)
	milk, err := b.squeeze(0)
	if err != nil {
		fmt.Printf("err = %s\n", err)
	}
	if milk {
		fmt.Printf("Milky milky!\n")
	}

	fmt.Printf("nipples = >>%#v<<\n", b)

	b2 := &Boobies{
		Size:    42,
		Nipples: 2,
	}
	b2.Shade.Colour = "Ginger"
	b2.Shade.Freckles = true
	b2.squeeze(1)

	titBits := map[string]*Boobies{}
	titBits["JennyTits"] = b
	titBits["NeechiTits"] = b2
	for name, value := range titBits {
		fmt.Printf("%v => %#v\n", name, value)
	}

}
