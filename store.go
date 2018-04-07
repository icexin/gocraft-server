package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"log"
	"math"
	"strconv"
	"time"

	"github.com/boltdb/bolt"
)

var (
	dbpath = flag.String("db", "gocraft.db", "db file name")
)

var (
	blockBucket  = []byte("block")
	chunkBucket  = []byte("chunk")
	cameraBucket = []byte("camera")

	store *Store
)

func InitStore() error {
	if *dbpath == "" {
		return nil
	}
	var err error
	store, err = NewStore(*dbpath)
	return err
}

type Store struct {
	db *bolt.DB
}

func NewStore(p string) (*Store, error) {
	db, err := bolt.Open(p, 0666, nil)
	if err != nil {
		return nil, err
	}
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(blockBucket)
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists(chunkBucket)
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists(cameraBucket)
		return err
	})
	if err != nil {
		return nil, err
	}
	db.NoSync = true
	return &Store{
		db: db,
	}, nil
}

func (s *Store) UpdateBlock(id Vec3, w int) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		log.Printf("put %v -> %d", id, w)
		bkt := tx.Bucket(blockBucket)
		cid := id.Chunkid()
		key := encodeBlockDbKey(cid, id)
		value := encodeBlockDbValue(w)
		return bkt.Put(key, value)
	})
}

func (s *Store) UpdateCamera(x, y, z, rx, ry float32) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(cameraBucket)
		buf := new(bytes.Buffer)
		binary.Write(buf, binary.LittleEndian, [...]float32{x, y, z, rx, ry})
		bkt.Put(cameraBucket, buf.Bytes())
		return nil
	})
}

func (s *Store) GetCamera() (x, y, z, rx, ry float32) {
	y = 16
	s.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(cameraBucket)
		value := bkt.Get(cameraBucket)
		if value == nil {
			return nil
		}
		buf := bytes.NewBuffer(value)
		binary.Read(buf, binary.LittleEndian, &x)
		binary.Read(buf, binary.LittleEndian, &y)
		binary.Read(buf, binary.LittleEndian, &z)
		binary.Read(buf, binary.LittleEndian, &rx)
		binary.Read(buf, binary.LittleEndian, &ry)
		return nil
	})
	return
}

func (s *Store) RangeBlocks(id Vec3, f func(bid Vec3, w int)) error {
	return s.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(blockBucket)
		startkey := encodeBlockDbKey(id, Vec3{0, 0, 0})
		iter := bkt.Cursor()
		for k, v := iter.Seek(startkey); k != nil; k, v = iter.Next() {
			cid, bid := decodeBlockDbKey(k)
			if cid != id {
				break
			}
			w := decodeBlockDbValue(v)
			f(bid, w)
		}
		return nil
	})
}

func (s *Store) UpdateChunkVersion(id Vec3, version string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(chunkBucket)
		key := encodeVec3(id)
		return bkt.Put(key, []byte(version))
	})
}

func (s *Store) GetChunkVersion(id Vec3) string {
	var version string
	s.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(chunkBucket)
		key := encodeVec3(id)
		v := bkt.Get(key)
		if v != nil {
			version = string(v)
		}
		return nil
	})
	return version
}

func (s *Store) Close() {
	s.db.Sync()
	s.db.Close()
}

func encodeVec3(v Vec3) []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, [...]int32{int32(v.X), int32(v.Y), int32(v.Z)})
	return buf.Bytes()
}

func encodeBlockDbKey(cid, bid Vec3) []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, [...]int32{int32(cid.X), int32(cid.Z)})
	binary.Write(buf, binary.LittleEndian, [...]int32{int32(bid.X), int32(bid.Y), int32(bid.Z)})
	return buf.Bytes()
}

func decodeBlockDbKey(b []byte) (Vec3, Vec3) {
	if len(b) != 4*5 {
		log.Panicf("bad db key length:%d", len(b))
	}
	buf := bytes.NewBuffer(b)
	var arr [5]int32
	binary.Read(buf, binary.LittleEndian, &arr)

	cid := Vec3{int(arr[0]), 0, int(arr[1])}
	bid := Vec3{int(arr[2]), int(arr[3]), int(arr[4])}
	if bid.Chunkid() != cid {
		log.Panicf("bad db key: cid:%v, bid:%v", cid, bid)
	}
	return cid, bid
}

func encodeBlockDbValue(w int) []byte {
	value := make([]byte, 4)
	binary.LittleEndian.PutUint32(value, uint32(w))
	return value
}

func decodeBlockDbValue(b []byte) int {
	if len(b) != 4 {
		log.Panicf("bad db value length:%d", len(b))
	}
	return int(binary.LittleEndian.Uint32(b))
}

func GenrateChunkVersion() string {
	return strconv.FormatInt(time.Now().UnixNano(), 16)
}

const (
	ChunkWidth = 32
)

type Vec3 struct {
	X, Y, Z int
}

func (v Vec3) Left() Vec3 {
	return Vec3{v.X - 1, v.Y, v.Z}
}
func (v Vec3) Right() Vec3 {
	return Vec3{v.X + 1, v.Y, v.Z}
}
func (v Vec3) Up() Vec3 {
	return Vec3{v.X, v.Y + 1, v.Z}
}
func (v Vec3) Down() Vec3 {
	return Vec3{v.X, v.Y - 1, v.Z}
}
func (v Vec3) Front() Vec3 {
	return Vec3{v.X, v.Y, v.Z + 1}
}
func (v Vec3) Back() Vec3 {
	return Vec3{v.X, v.Y, v.Z - 1}
}
func (v Vec3) Chunkid() Vec3 {
	return Vec3{
		int(math.Floor(float64(v.X) / ChunkWidth)),
		0,
		int(math.Floor(float64(v.Z) / ChunkWidth)),
	}
}
