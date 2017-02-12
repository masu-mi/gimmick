package s1

// Equal returns two elements' equality.
func Equal(a, b uint64) bool {
	return a == b
}

// Representative returns representative of attribute value.
type Representative func(uint64) uint64

// Equal returns Equality between attributes' representatives.
func (r Representative) Equal(a, b uint64) bool {
	if r == nil {
		return Equal(a, b)
	}
	return r(a) == r(b)
}

// Lift returns lift
func (r Representative) RotationNumber(points ...uint64) (lift int) {
	if r == nil {
		return RotationNumber(points...)
	}
	l := len(points)
	if l == 0 {
		return 0
	}
	for i := 1; i < l; i++ {
		if r(points[i-1]) > r(points[i]) {
			lift++
		}
	}
	if r(points[l-1]) > r(points[0]) {
		lift++
	}
	return lift
}

// RotationNumber returns value which described follow sentence.
// When S^1 projected to abs(z) = 1 in complex plane.
// log functions difference of start and end points over loop path(C_loop)
// which is coresponding to given attribute points.
// NOTE
//    C_loop:
//        points[0] -> points[1] -> ... ->
//        points[len(points)-1] =(return to start point)=> points[0]
//   Ex)
//        windingNum := 1: 1           => 1
//        windingNum := 1: 1 -> 1      => 1
//        windingNum := 1: 1 -> 2      => 1
//        windingNum := 1: 1 -> 3      => 1
//        windingNum := 2: 1 -> 3 -> 2 => 1
//        windingNum := 1: 1 -> 3 -> 3 => 1
func RotationNumber(points ...uint64) (lift int) {
	l := len(points)
	if l == 0 {
		return 0
	}
	for i := 1; i < l; i++ {
		if points[i-1] > points[i] {
			lift++
		}
	}
	if points[l-1] > points[0] {
		lift++
	}
	return lift
}
