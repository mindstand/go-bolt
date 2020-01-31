package types

const (
	Point2DStructSignature byte = 'X'
	Point2DStructSize      int  = 3

	Point3DStructSignature byte = 'Y'
	Point3DStructSize      int  = 4
)

type Point2D struct {
	SRID int
	X, Y float64
}

func (p *Point2D) Signature() int {
	return int(Point2DStructSignature)
}

func (p *Point2D) AllFields() []interface{} {
	return []interface{}{p.SRID, p.X, p.Y}
}

type Point3D struct {
	SRID    int
	X, Y, Z float64
}

func (p *Point3D) Signature() int {
	return int(Point3DStructSignature)
}

func (p *Point3D) AllFields() []interface{} {
	return []interface{}{p.SRID, p.X, p.Y, p.Z}
}
