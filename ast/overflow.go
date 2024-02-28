package fixtures

type ziran int64

func cacu() {
	var a, b int32 = 10, 20
	const x int32 = 99
	var m, n ziran = 102, 51

	sum := -a
	for i := 0; i <= 10; i++ {
		sum += i
	}
	c := a + b
	e := m << 3
	d := m - (c * n)
	m <<= 2
	sum = m % a
	n++
}
