package wasm

// The idea here is that we provide log and stuff here

// This function is imported from JavaScript, as it doesn't define a body.
// You should define a function named 'main.add' in the WebAssembly 'env'
// module from JavaScript.
func Log(string)
