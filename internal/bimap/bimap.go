package bimap

type BiMap[T1 comparable, T2 comparable] struct {
	m1 map[T1]T2
	m2 map[T2]T1
}

func New[T1 comparable, T2 comparable](cap int) *BiMap[T1, T2] {
	m1 := make(map[T1]T2, cap)
	m2 := make(map[T2]T1, cap)

	return &BiMap[T1, T2]{
		m1: m1,
		m2: m2,
	}
}

func (bm *BiMap[T1, T2]) GetByKey(key T1) (T2, bool) {
	v, ok := bm.m1[key]
	return v, ok
}

func (bm *BiMap[T1, T2]) GetByValue(value T2) (T1, bool) {
	k, ok := bm.m2[value]
	return k, ok
}

func (bm *BiMap[T1, T2]) DeleteByKey(key T1) {
	value, ok := bm.m1[key]
	if !ok {
		return
	}
	delete(bm.m1, key)
	delete(bm.m2, value)
}

func (bm *BiMap[T1, T2]) DeleteByValue(value T2) {
	key, ok := bm.m2[value]
	if !ok {
		return
	}
	delete(bm.m2, value)
	delete(bm.m1, key)
}

func (bm *BiMap[T1, T2]) Set(key T1, value T2) {
	bm.m1[key] = value
	bm.m2[value] = key
}

func (bm *BiMap[T1, T2]) Len() int {
	return len(bm.m1)
}

func (bm *BiMap[T1, T2]) Keys() []T1 {
	res := make([]T1, len(bm.m1))
	i := 0
	for k := range bm.m1 {
		res[i] = k
		i += 1
	}
	return res
}

func (bm *BiMap[T1, T2]) Values() []T2 {
	res := make([]T2, len(bm.m2))
	i := 0
	for v := range bm.m2 {
		res[i] = v
		i += 1
	}
	return res
}

func (bm *BiMap[T1, T2]) CopyM1() map[T1]T2 {
	r := make(map[T1]T2, len(bm.m1))
	for k, v := range bm.m1 {
		r[k] = v
	}
	return r
}
