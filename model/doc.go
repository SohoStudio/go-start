/*
The Model of MVC.

Models are regular Go structs. Reflection is used to marshall and unmarshall
them to and from other APIs.

The struct fields are of stock Go types except for some special types
that don't have a stock Go representation.
All special field types implement the Field interface.

Meta information for display and validation of values can be added via
Go struct field tags with the prefix `gostart:`.

WalkStructure() is utilized for validation an all other per-struct-field
operations. It provides a callback function with MetaData for every
struct field gathered by reflection.


*/
package model
