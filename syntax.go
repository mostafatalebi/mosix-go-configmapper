package configmapper

type Syntax = string

const (
	SyntaxArrayInt       Syntax = "array.int::"
	SyntaxArrayFloat     Syntax = "array.float::"
	SyntaxArrayStr       Syntax = "array.string::"
	SyntaxJsonObject     Syntax = "json.object::"
	SyntaxBase64Encoding Syntax = "base64.encode::"
	SyntaxBase64Decoding Syntax = "base64.decode::"
	SyntaxTimeDuration   Syntax = "time.duration::"
	SyntaxDataSize       Syntax = "data.size::"
	SyntaxURLEncode      Syntax = "url.encode::"
	SyntaxURLDecode      Syntax = "url.decode::"
	SyntaxURLParse       Syntax = "url::"
)
