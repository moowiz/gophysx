package gophysx

import (
	"github.com/stretchr/testify/require"
	"math/rand"
	"testing"
	"time"
)

type TestClock struct {
	currentTime time.Time
}

func (t *TestClock) Now() time.Time {
	return t.currentTime
}

func (t *TestClock) AddTime(d time.Duration) {
	t.currentTime = t.currentTime.Add(d)
}

func initSystem() (*System, *TestClock) {
	clock := TestClock{time.Now()}
	return Init(&clock), &clock
}

func makeTestObject(num_polygons int32) ([]Vector, Vector) {
	var polys []Vector
	for ; num_polygons > 0; num_polygons-- {
		polys = append(polys, Vector{rand.Float64(), rand.Float64()})
	}
	position := Vector{0, 0}
	return polys, position
}

func TestInitCollisionSystem(t *testing.T) {
	system, _ := initSystem()
	require.NotNil(t, system, "Init returned a nil value.")
	require.Empty(t, system.objects, "Collision system should start empty.") // TODO change?
}

func TestGetObjectEmpty(t *testing.T) {
	system, _ := initSystem()
	_, err := system.GetObject(0)
	require.NotNil(t, err, "Getting a non existant object should be nil")
}

func TestObjectBasics(t *testing.T) {
	system, _ := initSystem()
	polys, position := makeTestObject(4)

	obj, err := system.AddObject(polys, position)
	require.Nil(t, err, "Add object failed")
	require.NotNil(t, obj, "Didn't return a valid object")

	require.Equal(t, position, obj.Position(), "Positions are not the same")
	require.Equal(t, Vector{0, 0}, obj.Velocity(), "Objects should start at rest")
	obj2, err := system.GetObject(obj.Id())
	require.Nil(t, err, "Get object failed")
	require.NotNil(t, obj2, "Object should not be nil")
	require.Equal(t, *obj, *obj2, "Objects should be equal")

	err = obj.Remove()
	require.Nil(t, err, "Remove object failed")

	_, err = system.GetObject(obj.Id())
	require.NotNil(t, err, "Shouldn't be able to see the object anymore")
}

func TestForceBasics(t *testing.T) {
	system, _ := initSystem()
	polys, position := makeTestObject(4)
	obj, _ := system.AddObject(polys, position)

	magnitude := 10.0
	direction := Vector{1, 1}

	f, err := obj.AddForce(magnitude, direction)
	require.Equal(t, magnitude, f.Magnitude())
	require.Equal(t, direction.Normalize(), f.Direction())

	err = f.SetMagnitude(5.0)
	require.Nil(t, err, "Updating a force failed")
	f, err = obj.GetForce(f.Id()) // The value gets saved
	require.Nil(t, err, "Should be allowed to get this force")
	require.Equal(t, 5.0, f.Magnitude(), "Magnitude not updated")

	err = f.SetDirection(Vector{5.0, 5.0})
	require.Nil(t, err, "Updating a force failed")
	f, err = obj.GetForce(f.Id()) // The value gets saved
	require.Nil(t, err, "Should be allowed to get this force")
	require.Equal(t, Vector{5.0, 5.0}.Normalize(), f.Direction())

	require.Nil(t, f.Remove(), "Should be allowed to remove a force")

	_, err = obj.GetForce(f.Id())
	require.NotNil(t, err, "Should no longer be the force on the object")
}

func TestForceOverTime(t *testing.T) {
	system, clock := initSystem()
	polys, position := makeTestObject(4)
	position = position.Add(Vector{1, 1})
	obj, _ := system.AddObject(polys, position)

	magnitude := 1.0
	direction := Vector{1, 0}

	position = obj.Position()
	require.Equal(t, Vector{1, 1}, position, "Haven't moved yet")
	velocity := obj.Velocity()
	require.Equal(t, Vector{}, velocity, "Haven't moved yet")

	f, _ := obj.AddForce(magnitude, direction)
	clock.AddTime(time.Second)

	newVel := obj.Velocity()
	require.Equal(t, Vector{1, 0}, newVel.Sub(velocity), "Position didn't change")

	newPos := obj.Position()
	require.Equal(t, Vector{0.5, 0}, newPos.Sub(position), "Position didn't change")

	newPos2 := obj.Position()
	require.Equal(t, newPos, newPos2, "Position shouldn't change after calling a second time")

	newVel2 := obj.Velocity()
	require.Equal(t, newVel, newVel2, "Velocity shouldn't change after calling a second time")

	f.Remove()

	clock.AddTime(time.Second)

	newVel = obj.Velocity()
	require.Equal(t, newVel, newVel2, "Velocity stays the same when no force")

	newPos = obj.Position()
	require.Equal(t, newPos, newPos2, "Position stays the same when no force")
}
