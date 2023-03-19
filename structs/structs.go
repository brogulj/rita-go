package structs

type VisionApiResponse struct {
	TextAnnotations []TextAnnotation
}

type TextAnnotation struct {
	Description  string
	BoundingPoly BoundingPoly
}

type BoundingPoly struct {
	Vertices []Vertex
}

type Vertex struct {
	X float64
	Y float64
}

type Line struct {
	Start Vertex
	End   Vertex
}

type WordWithCoords struct {
	Vertices     []Vertex
	LeftBoundry  Line
	RightBoundry Line
	TopLine      Line
	BottomLine   Line
	MiddleLine   Line
	Center       Vertex
	Word         string
	Angle        float64
}

type WordMatch struct {
	FirstWordID  int
	SecondWordID int
	Distance     float64
	MiddleLine   bool
	TopLine      bool
	BottomLine   bool
}
