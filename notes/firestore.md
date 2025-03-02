# Reference
```go

      // Read from the "skeets" collection
      readSkeets(client)


        //Write to the "skeets" collection
        writeSkeet(client)


	//Update the document
	result, err := updateSkeetContent(client, "4inWued0MOoMjgBKK1Jy")
	if err != nil {
		log.Fatalf("Error updating skeet: %v", err)
	}
	if result != nil {
		fmt.Printf("Write result: %+v\n", result)
	}

    // delete
      result, err := deleteSkeet(client, "AViGIbBHnYTA31pKvJS7")
        if err != nil {
                log.Fatalf("Error deleting skeet: %v", err)
        }

        if result != nil{
                fmt.Printf("Write result: %+v\n",result)
        }


```
