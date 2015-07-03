package gophysx

/*
This is a 2D collision system.
*/

import (
	"errors"
	"time"
)

// For mocking
type Vector struct {
	x, y float64
}

func Vec(x, y float64) Vector {
	return Vector{x, y}
}

func (p Vector) X() float64 { return p.x }
func (p Vector) Y() float64 { return p.y }
func (p Vector) Normalize() Vector {
	mag := p.x*p.x + p.y*p.y
	if mag == 0 || mag == 1 {
		return p
	}
	return Vector{p.x / mag, p.y / mag}
}

func (p Vector) Add(p2 Vector) Vector {
	return Vector{p.x + p2.X(), p.y + p2.Y()}
}

func (p Vector) Scale(val float64) Vector {
	return Vector{p.x * val, p.y * val}
}

func (p Vector) Sub(p2 Vector) Vector {
	return p.Add(p2.Scale(-1))
}

func (p Vector) Mul(p2 Vector) Vector {
	return Vector{p.x * p2.x, p.y * p2.y}
}

type PhysxObj struct {
	id       int32
	polygon  []Vector
	position Vector
	velocity Vector
	forces   map[int32]*Force
	nextFid  int32
	system   *System
}

func (o *PhysxObj) Id() int32 { return o.id }

func (o *PhysxObj) recompute() {
	now := o.system.clock.Now()

	for _, value := range o.forces {
		delta := now.Sub(value.startTime)
		amt := value.direction.Scale(value.magnitude).Scale(delta.Seconds())

		o.position = o.position.Add(o.velocity.Add(amt.Mul(amt).Scale(.5)))
		o.velocity = o.velocity.Add(amt)
		value.startTime = now
	}
}

func (o *PhysxObj) Position() Vector {
	o.recompute()
	return o.position
}

func (o *PhysxObj) Velocity() Vector {
	o.recompute()
	return o.velocity
}

type Force struct {
	magnitude float64
	direction Vector
	startTime time.Time
	id        int32
	obj       *PhysxObj
}

func (f *Force) Magnitude() float64 { return f.magnitude }
func (f *Force) Direction() Vector  { return f.direction }
func (f *Force) Id() int32          { return f.id }

type Clock interface {
	Now() time.Time
}

type System struct {
	objects map[int32]*PhysxObj
	nextid  int32
	clock   Clock
}

func Init(clock Clock) *System {
	return &System{make(map[int32]*PhysxObj), 0, clock}
}

func (s *System) GetObjectPosition(id int32) (Vector, error) {
	obj, ok := s.objects[id]
	if !ok {
		return Vector{}, errors.New("Object does not exist")
	}
	return obj.position, nil
}

func (s *System) AddObject(polygon []Vector, position Vector) (*PhysxObj, error) {
	id := s.nextid
	s.nextid++

	if _, ok := s.objects[id]; ok {
		return nil, errors.New("Object already exists. This shouldn't happen ever...")
	}

	s.objects[id] = &PhysxObj{id, polygon, position, Vector{0, 0}, make(map[int32]*Force), 0, s}
	return s.objects[id], nil
}

func (s *System) GetObject(id int32) (*PhysxObj, error) {
	obj, ok := s.objects[id]
	if !ok {
		return nil, errors.New("Object does not exist")
	}
	return obj, nil
}

func (s *System) removeObject(id int32) error {
	if _, ok := s.objects[id]; !ok {
		return errors.New("Object does not exist")
	}
	delete(s.objects, id)
	return nil
}

func (o *PhysxObj) Remove() error {
	return o.system.removeObject(o.id)
}

/*
func (s *System) AddForce(id int32, fid int32) error {
	obj, ok := s.objects[id]
	if !ok {
		return errors.New("Object does not exist")
	}
	return obj
}*/

func (o *PhysxObj) AddForce(magnitude float64, direction Vector) (*Force, error) {
	fid := o.nextFid
	o.nextFid++

	if _, ok := o.forces[fid]; ok {
		return nil, errors.New("Force already exists... This shouldn't happen")
	}

	o.forces[fid] = &Force{magnitude, direction.Normalize(), o.system.clock.Now(), fid, o}
	return o.forces[fid], nil
}

func (o *PhysxObj) GetForce(id int32) (*Force, error) {
	obj, ok := o.forces[id]
	if !ok {
		return nil, errors.New("Force does not exist")
	}
	return obj, nil
}

func (o *PhysxObj) removeForce(fid int32) error {
	if _, ok := o.forces[fid]; !ok {
		return errors.New("Force does not exist")
	}
	o.recompute()
	delete(o.forces, fid)
	return nil
}

func (f *Force) Remove() error {
	return f.obj.removeForce(f.id)
}

func (f *Force) SetMagnitude(value float64) error {
	f.obj.recompute()
	f.magnitude = value
	return nil
}

func (f *Force) SetDirection(value Vector) error {
	f.obj.recompute()
	f.direction = value.Normalize()
	return nil
}
