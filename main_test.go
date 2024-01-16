package main

import (
	"fmt"
	"testing"
)

func TestOrganiseText(t *testing.T) {
	const text = "some text here\n" +
		"```\n" +
		"print(\"some code\")\n" +
		"```\n" +
		"some more text"

	// fmt.Println(text)
	ans := organiseText(text)
	fmt.Println(ans[2])
	// fmt.Println(ans)
	if len(ans) != 3 {
		t.Errorf("organiseText(text) = %v, want %v", len(ans), 3)
	}
}
