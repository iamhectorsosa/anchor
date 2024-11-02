package database_test

import (
	"testing"

	"github.com/iamhectorsosa/anchor/internal/database"
	"github.com/iamhectorsosa/anchor/internal/store"
)

func TestCreateAndRead(t *testing.T) {
	db, cleanup, err := database.NewInMemory()
	if err != nil {
		t.Fatalf("database.NewInMemory, err=%v", err)
	}
	defer cleanup()

	want := store.Anchor{
		Key:   "hello",
		Value: "world",
	}

	if err := db.Create(want.Key, want.Value); err != nil {
		t.Errorf("db.Create, err=%v", err)
	}

	got, err := db.Read(want.Key)
	if err != nil {
		t.Errorf("db.Read, err=%v", err)
	}

	if got.Key != want.Key || got.Value != want.Value {
		t.Errorf("unexpected key values, wanted %s=%q got %s=%q", want.Key, want.Value, got.Key, got.Value)
	}
}

func TestCreateAndReadAll(t *testing.T) {
	db, cleanup, err := database.NewInMemory()
	if err != nil {
		t.Fatalf("database.NewInMemory, err=%v", err)
	}
	defer cleanup()

	wantAnchors := []store.Anchor{
		store.Anchor{
			Key:   "hello",
			Value: "world",
		},
		store.Anchor{
			Key:   "ahoj",
			Value: "ciao",
		},
	}

	for _, s := range wantAnchors {
		if err := db.Create(s.Key, s.Value); err != nil {
			t.Errorf("db.Create, err=%v", err)
		}
	}

	anchors, err := db.ReadAll()
	if err != nil {
		t.Errorf("db.Read, err=%v", err)
	}

	if len(anchors) != 2 {
		t.Errorf("unexpected number of anchors, wanted 2 got %d", len(anchors))
	}

	for i, got := range anchors {
		want := wantAnchors[i]
		if got.Key != want.Key || got.Value != want.Value {
			t.Errorf("unexpected key values, wanted %s=\"%q\" got %s=%q", want.Key, want.Value, got.Key, got.Value)
		}
	}
}

func TestCreateUpdateAndRead(t *testing.T) {
	db, cleanup, err := database.NewInMemory()
	if err != nil {
		t.Fatalf("database.NewInMemory, err=%v", err)
	}
	defer cleanup()

	want := store.Anchor{
		Key:   "hello",
		Value: "world",
	}

	if err := db.Create(want.Key, want.Value); err != nil {
		t.Errorf("db.Create, err=%v", err)
	}

	got, err := db.Read(want.Key)
	if err != nil {
		t.Errorf("db.Read, err=%v", err)
	}

	if got.Key != want.Key || got.Value != want.Value {
		t.Errorf("unexpected key values, wanted %s=%q got %s=%q", want.Key, want.Value, got.Key, got.Value)
	}

	updateWant := store.Anchor{
		Key:   "hello",
		Value: "goodbye",
	}

	err = db.Update(updateWant)
	if err != nil {
		t.Errorf("db.Update, err=%v", err)
	}

	got, err = db.Read(updateWant.Key)
	if err != nil {
		t.Errorf("db.Read, err=%v", err)
	}

	if got.Key != updateWant.Key || got.Value != updateWant.Value {
		t.Errorf("unexpected key values, wanted %s=%q got %s=%q", updateWant.Key, updateWant.Value, got.Key, got.Value)
	}
}

func TestCreateDeleteAndRead(t *testing.T) {
	db, cleanup, err := database.NewInMemory()
	if err != nil {
		t.Fatalf("database.NewInMemory, err=%v", err)
	}
	defer cleanup()

	want := store.Anchor{
		Key:   "hello",
		Value: "world",
	}

	if err := db.Create(want.Key, want.Value); err != nil {
		t.Errorf("db.Create, err=%v", err)
	}

	got, err := db.Read(want.Key)
	if err != nil {
		t.Errorf("db.Read, err=%v", err)
	}

	if got.Key != want.Key || got.Value != want.Value {
		t.Errorf("unexpected key values, wanted %s=%q got %s=%q", want.Key, want.Value, got.Key, got.Value)
	}

	err = db.Delete(want.Key)
	if err != nil {
		t.Errorf("db.Delete, err=%v", err)
	}

	if _, err := db.Read(want.Key); err == nil {
		t.Errorf("expected error err=sql: no rows in result set")
	}
}

func TestCreateResetAndReadAll(t *testing.T) {
	db, cleanup, err := database.NewInMemory()
	if err != nil {
		t.Fatalf("database.NewInMemory, err=%v", err)
	}
	defer cleanup()

	wantAnchors := []store.Anchor{
		store.Anchor{
			Key:   "hello",
			Value: "world",
		},
		store.Anchor{
			Key:   "ahoj",
			Value: "ciao",
		},
	}

	for _, s := range wantAnchors {
		if err := db.Create(s.Key, s.Value); err != nil {
			t.Errorf("db.Create, err=%v", err)
		}
	}

	anchors, err := db.ReadAll()
	if err != nil {
		t.Errorf("db.Read, err=%v", err)
	}

	if len(anchors) != 2 {
		t.Errorf("unexpected number of anchors, wanted 2 got %d", len(anchors))
	}

	for i, got := range anchors {
		want := wantAnchors[i]
		if got.Key != want.Key || got.Value != want.Value {
			t.Errorf("unexpected key values, wanted %s=\"%q\" got %s=%q", want.Key, want.Value, got.Key, got.Value)
		}
	}

	if err := db.Reset(); err != nil {
		t.Errorf("db.Reset, err=%v", err)
	}

	anchors, err = db.ReadAll()
	if err != nil {
		t.Errorf("db.Read, err=%v", err)
	}

	if len(anchors) != 0 {
		t.Errorf("unexpected number of anchors, wanted 0 got %d", len(anchors))
	}
}

func TestImportAndReadAll(t *testing.T) {
	db, cleanup, err := database.NewInMemory()
	if err != nil {
		t.Fatalf("database.NewInMemory, err=%v", err)
	}
	defer cleanup()

	wantAnchors := []store.Anchor{
		store.Anchor{
			Key:   "hello",
			Value: "world",
		},
		store.Anchor{
			Key:   "ahoj",
			Value: "ciao",
		},
	}

	if err := db.Import(wantAnchors); err != nil {
		t.Errorf("db.Create, err=%v", err)
	}

	anchors, err := db.ReadAll()
	if err != nil {
		t.Errorf("db.Read, err=%v", err)
	}

	if len(anchors) != 2 {
		t.Errorf("unexpected number of anchors, wanted 2 got %d", len(anchors))
	}

	for i, got := range anchors {
		want := wantAnchors[i]
		if got.Key != want.Key || got.Value != want.Value {
			t.Errorf("unexpected key values, wanted %s=\"%q\" got %s=%q", want.Key, want.Value, got.Key, got.Value)
		}
	}
}
