package loaders

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

// LoadObj now returns a map: MaterialName -> VertexData
func LoadObj(filename string) map[string][]float32 {
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	var vertices [][3]float32
	var textures [][2]float32
	var normals [][3]float32

	// The output map
	finalData := make(map[string][]float32)

	// Default material name if none is provided
	currentMaterial := "default"

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}

		switch parts[0] {
		case "mtllib":
			// We ignore the .mtl file for now and handle textures manually in main.go
			continue
		case "usemtl":
			// SWITCH MATERIAL!
			// Future faces will be added to this key
			currentMaterial = parts[1]
		case "v":
			x, _ := strconv.ParseFloat(parts[1], 32)
			y, _ := strconv.ParseFloat(parts[2], 32)
			z, _ := strconv.ParseFloat(parts[3], 32)
			vertices = append(vertices, [3]float32{float32(x), float32(y), float32(z)})
		case "vt":
			u, _ := strconv.ParseFloat(parts[1], 32)
			v, _ := strconv.ParseFloat(parts[2], 32)
			textures = append(textures, [2]float32{float32(u), float32(v)})
		case "vn":
			x, _ := strconv.ParseFloat(parts[1], 32)
			y, _ := strconv.ParseFloat(parts[2], 32)
			z, _ := strconv.ParseFloat(parts[3], 32)
			normals = append(normals, [3]float32{float32(x), float32(y), float32(z)})
		case "f":
			for i := 1; i <= 3; i++ {
				subParts := strings.Split(parts[i], "/")

				vIdx, _ := strconv.Atoi(subParts[0])
				v := vertices[vIdx-1]

				tIdx, _ := strconv.Atoi(subParts[1])
				t := textures[tIdx-1]

				nIdx, _ := strconv.Atoi(subParts[2])
				n := normals[nIdx-1]

				// Append to the SPECIFIC slice for the current material
				finalData[currentMaterial] = append(finalData[currentMaterial],
					v[0], v[1], v[2],
					n[0], n[1], n[2],
					t[0], t[1],
				)
			}
		}
	}
	return finalData
}
