package isodb

type (
	inMemBlobMap struct {
		items map[BlobRef]Blob
	}

	kvBlobMap struct {
		kv    KV
		cache blobMap
	}
)

func (bm *inMemBlobMap) put(b toBlober) {
	bm.ensureItems()
	bm.items[b.ToBlob().Ref()] = b.ToBlob()
}

func (bm *inMemBlobMap) read(out interface{}, r BlobRef) bool {
	bm.ensureItems()
	v, ok := bm.items[r]
	if !ok {
		return ok
	}
	err := defaultCodec.decode(out, v)
	if err != nil {
		panic(err)
	}
	return true
}

func (bm *inMemBlobMap) ensureItems() {
	if bm.items == nil {
		bm.items = make(map[BlobRef]Blob)
	}
}

func (bm *inMemBlobMap) has(r BlobRef) bool {
	bm.ensureItems()
	_, ok := bm.items[r]
	return ok
}

func (bm *inMemBlobMap) keys() []BlobRef {
	ret := make([]BlobRef, 0, len(bm.items))
	for k := range bm.items {
		ret = append(ret, k)
	}
	return ret
}

func (bm *inMemBlobMap) raw(out []byte, r BlobRef) Blob {
	if !bm.has(r) {
		return Blob{}
	}
	b := bm.items[r]
	if len(b.Content) == 0 {
		return Blob{}
	}
	return Blob{Content: append(out, b.Content...)}
}

func (km *kvBlobMap) put(b toBlober) {
	km.cache.put(b)
}

func (km *kvBlobMap) has(b BlobRef) bool {
	if km.cache.has(b) {
		return true
	}
	has, err := km.kv.Has(b.String())
	if err != nil {
		panic(err)
	}
	return has
}

func (km *kvBlobMap) read(out interface{}, r BlobRef) bool {
	if !km.cache.has(r) {
		if !km.primeCache(r) {
			return false
		}
	}
	return km.cache.read(out, r)
}

func (km *kvBlobMap) raw(out []byte, r BlobRef) Blob {
	if !km.cache.has(r) {
		km.primeCache(r)
	}
	return km.cache.raw(out, r)
}

func (km *kvBlobMap) primeCache(r BlobRef) bool {
	has, err := km.kv.Has(r.String())
	if err != nil {
		panic(err)
	} else if !has {
		return false
	}

	blob, err := km.kv.Get(r.String())
	if err != nil {
		panic(err)
	}
	if len(blob.Content) == 0 {
		return false
	}
	km.cache.put(blob)
	return true
}

func (km *kvBlobMap) keys() []BlobRef {
	return km.cache.keys()
}
