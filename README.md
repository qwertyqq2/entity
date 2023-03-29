
## Entity ##  

#### Parallel data writing to storage using timings ####  

### Install ###

    go get github.com/qwertyqq2/entity

### Usage ###

    // Key-value store
    type Entity interface {
        //Closeing entity
        Shutdown()

        //Connect process to entity
        Connect(id int) (chan<- Message, error)

        //Disconnect a process from an entity
        Disconnect(id int) error

        //Getting response on request asynchronously
        Resp(id int) Message

        //Display data in entities
        String() string

        // Len of entity
        Len() int
    }

    // Interface that writes data to entity
	type Process interface {
		//Closing a process
		Shutdown()

		// Registration entity for process
		// Returns an error if the process is already registered or if the entity is closed
		Registration(ent entity.Entity) error

		//Adds data to the process, after which they will go to the entity
		Add(key, data string)

		//Creates a request to remove data from an entity
		Delete(key string)

		//Start a process
		Start(ent entity.Entity) error

		//Process id
		ID() int
	}

### Example ###

        // new entity with limit of count process
        entity := entity.New(limitProc)

        // number of process
        curNumber := 1

        // Create and run a process with options
        process.WithEntity(
            // number of process
            curNumber,
            // entity to write data
            entity,
            // interval after which data is sent
            process.ResentInterval(1*time.Second),
            // max size of message
            process.MaxMsgSize(256), //
            // interval after which a reconnection attempt is made
            process.ReconnectInterval(1*time.Second), //
            // if we have collected more data than this value,
            // then an attempt is made to send entity data
            process.SendMsgCutoff(100), //
            //response timeout from entity
            process.WaitRespInterval(5*time.Second),
            // if there is no connection during this interval,
            // then the process is closed
            process.MaxWaitingConnection(100*time.Second),
        )
