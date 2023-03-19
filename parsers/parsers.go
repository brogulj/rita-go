package visionToText

import (
	"math"
	structs "rita-go/structs"
)

func main(response structs.VisionApiResponse) {

}

func GetWordsWithCoords(response structs.VisionApiResponse) []structs.WordWithCoords {

	var wordsWithCoords []structs.WordWithCoords

	const LINE_LENGTH = 1000.0

	for i := 1; i < len(response.TextAnnotations); i++ {
		var vertices = response.TextAnnotations[i].BoundingPoly.Vertices
		var angle = math.Atan2(vertices[1].Y-vertices[0].Y, vertices[1].X-vertices[0].X)
		var wordWithCoords structs.WordWithCoords

		vertices = response.TextAnnotations[i].BoundingPoly.Vertices
		wordWithCoords.Word = response.TextAnnotations[i].Description

		wordWithCoords.LeftBoundry = structs.Line{
			Start: structs.Vertex{
				X: vertices[0].X,
				Y: vertices[0].Y,
			},
			End: structs.Vertex{
				X: vertices[3].X,
				Y: vertices[3].Y,
			},
		}
		wordWithCoords.RightBoundry = structs.Line{
			Start: structs.Vertex{
				X: vertices[1].X,
				Y: vertices[1].Y,
			},
			End: structs.Vertex{
				X: vertices[2].X,
				Y: vertices[2].Y,
			},
		}

		wordWithCoords.TopLine = structs.Line{
			Start: structs.Vertex{
				X: vertices[0].X,
				Y: vertices[0].Y,
			},
			End: structs.Vertex{
				X: vertices[1].X + LINE_LENGTH*math.Cos(angle),
				Y: vertices[1].Y + LINE_LENGTH*math.Sin(angle),
			},
		}
		wordWithCoords.BottomLine = structs.Line{
			Start: structs.Vertex{
				X: vertices[3].X,
				Y: vertices[3].Y,
			},
			End: structs.Vertex{
				X: vertices[2].X + LINE_LENGTH*math.Cos(angle),
				Y: vertices[2].Y + LINE_LENGTH*math.Sin(angle),
			},
		}
		wordWithCoords.MiddleLine = structs.Line{
			Start: structs.Vertex{
				X: vertices[0].X + (vertices[3].X-vertices[0].X)/2,
				Y: vertices[0].Y + (vertices[3].Y-vertices[0].Y)/2,
			},
			End: structs.Vertex{
				X: vertices[1].X + (vertices[1].X-vertices[2].X)/2 + LINE_LENGTH*math.Cos(angle),
				Y: vertices[1].Y + (vertices[1].Y-vertices[2].Y)/2 + LINE_LENGTH*math.Sin(angle),
			},
		}

		wordWithCoords.Center = structs.Vertex{
			X: (vertices[0].X +
				vertices[1].X +
				vertices[2].X +
				vertices[3].X) / 4.0,
			Y: (vertices[0].Y +
				vertices[1].Y +
				vertices[2].Y +
				vertices[3].Y) / 4.0,
		}

		wordWithCoords.Angle = angle
		wordsWithCoords = append(wordsWithCoords, wordWithCoords)
	}

	return wordsWithCoords
}

func GetMatches(words []structs.WordWithCoords) []structs.WordMatch {
	var wordMatches []structs.WordMatch = []structs.WordMatch{}

	for i := 0; i < len(words); i++ {
		var wordMatch = structs.WordMatch{
			FirstWordID:  i,
			SecondWordID: -1,
			Distance:     100000000,
			MiddleLine:   false,
			TopLine:      false,
			BottomLine:   false,
		}

		for j := 0; j < len(words); j++ {
			if i == j {
				continue
			}

			var firstWord = words[i]
			var secondWord = words[j]

			if firstWord.Center.X > secondWord.Center.X {
				continue
			}

			var topIntersect = doIntersect(firstWord.TopLine, secondWord.RightBoundry)
			var bottomIntersect = doIntersect(firstWord.BottomLine, secondWord.RightBoundry)
			var middleIntersect = doIntersect(firstWord.MiddleLine, secondWord.RightBoundry)

			if !topIntersect && !bottomIntersect && !middleIntersect {
				continue
			}

			var newDistance = math.Sqrt(math.Pow(firstWord.RightBoundry.Start.X-secondWord.LeftBoundry.Start.X, 2) + math.Pow(firstWord.RightBoundry.Start.Y-secondWord.LeftBoundry.Start.Y, 2))

			if newDistance < wordMatch.Distance {
				wordMatch = structs.WordMatch{
					FirstWordID:  i,
					SecondWordID: j,
					Distance:     newDistance,
					MiddleLine:   middleIntersect,
					TopLine:      topIntersect,
					BottomLine:   bottomIntersect,
				}
			}
		}
		wordMatches = append(wordMatches, wordMatch)
	}

	wordMatches = GetTrueSuccessor(wordMatches, words)

	return wordMatches
}

func GetTrueSuccessor(matches []structs.WordMatch, words []structs.WordWithCoords) []structs.WordMatch {
	for i := 0; i < len(matches); i++ {
		for j := 0; j < len(matches); j++ {
			if matches[i].SecondWordID != matches[j].SecondWordID || i == j || i > j {
				continue
			}

			if matches[i].SecondWordID == -1 {
				continue
			}

			if matches[i].Distance > matches[j].Distance {
				matches[i].SecondWordID = -1
			} else {
				matches[j].SecondWordID = -1
			}

		}
	}
	return matches
}

func BuildLines(matches []structs.WordMatch, words []structs.WordWithCoords) map[int]string {
	var lines []string
	var Lines = make(map[int]string)

	var startingWords []int = []int{}
	var startingMatches []structs.WordMatch = []structs.WordMatch{}

	for i := 0; i < len(words); i++ {
		startingWords = append(startingWords, i)
	}

	for i := 0; i < len(matches); i++ {
		var index = indexOf(matches[i].SecondWordID, startingWords)

		if index != -1 {
			startingWords = append(startingWords[:index], startingWords[index+1:]...)
		}
	}

	for i := 0; i < len(startingWords); i++ {
		for j := 0; j < len(matches); j++ {
			if matches[j].FirstWordID == startingWords[i] {
				startingMatches = append(startingMatches, matches[j])
			}
		}
	}

	for i := 0; i < len(startingMatches); i++ {
		var match = startingMatches[i]
		var line = words[match.FirstWordID].Word
		var nextWord = match.SecondWordID

		if nextWord == -1 {
			lines = append(lines, line)
			continue
		}

		for nextWord != -1 {
			line += " " + words[nextWord].Word
			nextWord = matches[nextWord].SecondWordID
		}
		Lines[int(words[match.FirstWordID].Center.Y)] = line
	}

	return Lines
}

func onSegment(p structs.Vertex, q structs.Vertex, r structs.Vertex) bool {
	if q.X <= math.Max(p.X, r.X) && q.X >= math.Min(p.X, r.X) && q.Y <= math.Max(p.Y, r.Y) && q.Y >= math.Min(p.Y, r.Y) {
		return true
	}
	return false
}

func orientation(p structs.Vertex, q structs.Vertex, r structs.Vertex) int {
	val := (q.Y-p.Y)*(r.X-q.X) - (q.X-p.X)*(r.Y-q.Y)
	if val == 0 {
		return 0
	}
	if val > 0 {
		return 1
	}
	return 2
}

func doIntersect(l1 structs.Line, l2 structs.Line) bool {
	o1 := orientation(l1.Start, l1.End, l2.Start)
	o2 := orientation(l1.Start, l1.End, l2.End)
	o3 := orientation(l2.Start, l2.End, l1.Start)
	o4 := orientation(l2.Start, l2.End, l1.End)

	if o1 != o2 && o3 != o4 {
		return true
	}

	if o1 == 0 && onSegment(l1.Start, l2.Start, l1.End) {
		return true
	}

	if o2 == 0 && onSegment(l1.Start, l2.End, l1.End) {
		return true
	}

	if o3 == 0 && onSegment(l2.Start, l1.Start, l2.End) {
		return true
	}

	if o4 == 0 && onSegment(l2.Start, l1.End, l2.End) {
		return true
	}

	return false
}

func indexOf(element int, data []int) int {
	for k, v := range data {
		if element == v {
			return k
		}
	}
	return -1 //not found.
}
