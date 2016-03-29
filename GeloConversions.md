For Gelo to Gelo conversions, see [API](API.md)

# Go to Gelo #

### Generic conversion ###
`Convert(interface{}) Word`

`Convert` is a very generic routine. It is meant for situations where we can have no advance type information, such as the [`Register\*` family](UsingTheVM#Adding_items_to_a_VM:_The_Register_*_family.md). It can convert to any type that the GeloCore recognizes.

If `Convert` is given a type that it is does not recognize it will panic with the [Error](GeloErrors#external_errors.md) "Convert given unknown type".

As a special case, calling `gelo.Convert(nil)` returns the value `Null`. If no conversion is possible `Convert` panics. `Convert` first attempts to coerce the value to a number via [NewNumberFromGo](GeloConversions#NumericConversions.md). Then `Convert` does the following conversions in the following order:

| Go type | Gelo type or Gelo conversion function called |
|:--------|:---------------------------------------------|
| `Word`  | `Word` (via `(Word).Copy() Word`)            |
| `func(*VM, *List, uint)` | `Alien`                                      |
| `bool`  | `gelo.ToBool(bool) Bool`                     |
| `string` | `Symbol`                                     |
| `[]byte` | `Symbol`                                     |
| `[]int` | `Symbol`                                     |
| `[]interface{}` | `*List`, each item in the slice is passed through `Convert` |
| `map[string]interface{}` | `*Dict`, each value in the map is passed through `Convert` |


### Numeric construction ###
`NewNumberFromGo(interface{}) (*Number, bool)`
> If `interface{}` is any of the Go numeric types (`int`, `float32`, etc), except the `complex` types, which cannot be represented in Gelo. If `interface{}` is `[]byte`, `string`, or `Word`, it calls `NewNumberFromBytes`, `NewNumberFromString`, or `NewNumberFrom`, respectively. Otherwise, it returns `(nil, false)`.

`NewNumberFromBytes([]byte) (*Number, bool)`

`NewNumberFromString(string) (*Number, bool)`
> Return `(nil, false)` if `strconv.Atof64(string)` returns false or the `*Number` constructed from the returned `float64` other.

`NewNumberFrom(Word) (*Number, bool)`
> Same as the above but calls `(Word).Ser().String()` first.

### Boolean construction ###
`ToBool(bool) Bool`

Maps:
| from  | to    |
|:------|:------|
|`true `|`True` |
|`false`|`False`|

### Symbol construction ###
`StrToSym(string) Symbol`

`BytesToSym([]byte) Symbol`

`RunesToSym([]int) Symbol`

### Quote construction ###
`NewQuoteFrom(Word) Quote`

`NewQuoteFromGo([]byte) Quote`


### Dict construction ###
`NewDictFrom(map[string]Word) *Dict`
> Copies all values.

`NewDictFromGo(map[string]interface{}) *Dict`
> Calls `Convert` on each value in the map. As such, can panic if given an unknown type.

`UnserializeDict([]byte, bool) (*Dict, bool)`
> Unserialize the `[]byte` slice if it is in the format that is produced by `(*Dict).Ser() Symbol`. The boolean paramater specifies whether the outermost `{}` are included. If the dictionary cannot be unserialized, returns `(nil, false)`.

`UnserializeDictFrom(Word) (*Dict, bool)`
> Same as above, but from `(Word).Ser().Bytes() []byte`. Only accepts `Symbol` or `Quote`.  All others return `(nil, false)`.

### List construction ###
`NewList(...Word) *List`
> Calls `NewListFrom`.

`NewListFrom([]Word) *List`
> Copies elements.

`NewListFromGo([]interface{}) *List`
> Calls `Convert` on each item in slice. As such, can panic if given an unknown type.

`UnserializeList([]byte, bool) (*List, bool)`
> Unserialize the `[]byte` slice if it is in the format that is produced by `(*List).Ser() Symbol`. The boolean paramater specifies whether the outermost `{}` are included. If the list cannot be unserialized, returns `(nil, false)`.

`UnserializeListFrom(Word) (*List, bool)`
> Same as above, but from `(Word).Ser().Bytes() []byte`. Only accepts `Symbol` or `Quote`.  All others return `(nil, false)`.

See also: ListBuilder

# Gelo to Go #

### Numeric conversions ###

`(*Number).Real() float64`

`(*Number).Int() (int64, bool)`
> The boolean return is false if the number cannot be represented as an `int64`.

### Boolean conversions ###

`(Bool).True() bool`

### String conversions ###
`(Symbol).String() string`

`(Symbol).Bytes() []byte`

`(Symbol).Runes() []int`

### [Dict](Dict.md) conversions ###

`(*Dict).Map() map[string]Word`
> Values are not copied.

## [List](List.md) conversions ##

`(*List).Slice() []Word`
> Elements are not copied.