package main

func main() {
	c := MakeCollector()

	c.Visit("https://xeiaso.net/blog/twitter-fears")

	c.Wait()
}
