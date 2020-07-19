package bast

//Pattern Pattern obj
type Pattern struct {
	Method        string
	Pattern       string
	Fn            func(ctx *Context)
	Parameter     interface{}
	Name          string
	Service       string
	authorization bool
	publish       bool
	publishFinish bool
	toRouter      bool
}

//Auth need api authorization
//eq Authorization
func (c *Pattern) Auth() *Pattern {
	c.authorization = true
	return c
}

//Unauth need api Unauthorization
//eq Unauthorization
func (c *Pattern) Unauth() *Pattern {
	c.authorization = false
	return c
}

//Publish register to etcd etc.
func (c *Pattern) Publish(service string) *Pattern {
	c.publish = true
	c.Service = service
	return c
}

//Unpublish unregister to etcd etc.
func (c *Pattern) Unpublish() *Pattern {
	c.publish = false
	c.Service = ""
	return c
}

//Param set pouter parameter
func (c *Pattern) Param(Parameter interface{}) *Pattern {
	c.Parameter = Parameter
	return c
}

//Nickname set nickname
func (c *Pattern) Nickname(name string) *Pattern {
	c.Name = name
	return c
}

//Router register to httpRouter
func (c *Pattern) Router() *Pattern {
	if !c.toRouter {
		doHandle(c)
		c.toRouter = true
	}
	return c
}
