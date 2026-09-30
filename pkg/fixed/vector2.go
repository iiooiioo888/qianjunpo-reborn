package fixed

// Vector2 is a 2D vector using fixed-point components.
type Vector2 struct {
	X, Y Fixed
}

func NewVector2(x, y Fixed) Vector2 {
	return Vector2{X: x, Y: y}
}

func (v Vector2) Add(o Vector2) Vector2 {
	return Vector2{X: v.X.Add(o.X), Y: v.Y.Add(o.Y)}
}

func (v Vector2) Sub(o Vector2) Vector2 {
	return Vector2{X: v.X.Sub(o.X), Y: v.Y.Sub(o.Y)}
}

func (v Vector2) Scale(s Fixed) Vector2 {
	return Vector2{X: v.X.Mul(s), Y: v.Y.Mul(s)}
}

// LengthSq returns squared length without sqrt.
func (v Vector2) LengthSq() Fixed {
	return v.X.Mul(v.X).Add(v.Y.Mul(v.Y))
}

// Length returns the Euclidean length using deterministic Sqrt.
func (v Vector2) Length() Fixed {
	return Sqrt(v.LengthSq())
}
