/*
	Author: Stefan Nilsson 2013-02-27

	Modified and optimized by Ivan Liljeqvist 2015-04-13.

	From the beginning the program executed linearly without concurrency.
	My task was to use go routines to utilize the power of all CPUs and make the program run faster.

	This program creates pictures of Julia sets (en.wikipedia.org/wiki/Julia_set).
*/

/*
	From the beginning the program executed in 25-26 seconds.
	With runtime.GOMAXPROCS(number_of_CPUs) the program executed in 25-26 seconds.
	With a go-routine for each pixel the program executed in 6-7 seconds.

	My program used 8 CPUs.
*/

package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"math/cmplx"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"
)

type ComplexFunc func(complex128) complex128

var Funcs []ComplexFunc = []ComplexFunc{
	func(z complex128) complex128 { return z*z - 0.61803398875 },
	func(z complex128) complex128 { return z*z + complex(0, 1) },
	func(z complex128) complex128 { return z*z + complex(-0.835, -0.2321) },
	func(z complex128) complex128 { return z*z + complex(0.45, 0.1428) },
	func(z complex128) complex128 { return z*z*z + 0.400 },
	func(z complex128) complex128 { return cmplx.Exp(z*z*z) - 0.621 },
	func(z complex128) complex128 { return (z*z+z)/cmplx.Log(z) + complex(0.268, 0.060) },
	func(z complex128) complex128 { return cmplx.Sqrt(cmplx.Sinh(z*z)) + complex(0.065, 0.122) },
}

func main() {
	//remeber when the program started
	start := time.Now()

	//produce the pictures
	//loop through all the functions
	for n, fn := range Funcs {
		//create the image by passing the function fn
		err := CreatePng("picture-"+strconv.Itoa(n)+".png", fn, 1024)
		//check for errors
		if err != nil {
			log.Fatal(err)
		}

	}

	//calculate the elapsed time
	elapsed := time.Since(start)

	//print the time it took for the program to execute
	fmt.Printf("The program executed in %s", elapsed)
}

// CreatePng creates a PNG picture file with a Julia image of size n x n.
func CreatePng(filename string, f ComplexFunc, n int) (err error) {
	//create the file that will hold the image
	file, err := os.Create(filename)
	//check for errors.
	if err != nil {
		return
	}
	//when evertyhing else in this method is finished - close file
	defer file.Close()
	//make the return variable by encoding the file using the function f and size n
	err = png.Encode(file, Julia(f, n))

	return
}

// Julia returns an image of size n x n of the Julia set for f.
func Julia(f ComplexFunc, n int) image.Image {
	//set up the image
	bounds := image.Rect(-n/2, -n/2, n/2, n/2)
	img := image.NewRGBA(bounds)

	s := float64(n / 4)

	//THIS IS THE HEAVY PART - we'll divide the process into routines

	wg := new(sync.WaitGroup)

	//we'll need a routine for each pixel - calculate the number of pixels
	image_width := (bounds.Max.X - bounds.Min.X)
	image_height := (bounds.Max.Y - bounds.Min.Y)
	number_of_pixels := image_height * image_width

	//we'll wait for each pixel to be calculated
	wg.Add(number_of_pixels)

	for i := bounds.Min.X; i < bounds.Max.X; i++ {
		for j := bounds.Min.Y; j < bounds.Max.Y; j++ {

			//create a routine to calculate this pixel

			go func(i, j int) {
				n := Iterate(f, complex(float64(i)/s, float64(j)/s), 256)

				r := uint8(0)
				g := uint8(0)
				//use n to generate the fractal
				b := uint8(n % 32 * 8)

				//save all the calculations to the image
				img.Set(i, j, color.RGBA{r, g, b, 255})

				wg.Done()
			}(i, j)

		}
	}

	//wait for all the pixels to be calculated
	wg.Wait()

	return img
}

// Iterate sets z_0 = z, and repeatedly computes z_n = f(z_{n-1}), n â‰¥ 1,
// until |z_n| > 2  or n = max and returns this n.
func Iterate(f ComplexFunc, z complex128, max int) (n int) {

	for ; n < max; n++ {
		if real(z)*real(z)+imag(z)*imag(z) > 4 {
			break
		}

		z = f(z)
	}
	return
}

func init() {
	//get number of CPUs
	number_of_CPUs := runtime.NumCPU()

	//attempt to use all CPUs
	runtime.GOMAXPROCS(number_of_CPUs)

	//print the number of CPUs
	fmt.Println("CPU´s: ", number_of_CPUs)

}
