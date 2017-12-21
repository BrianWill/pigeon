
// func main
//     locals file I err Str bytes S<Byte> n I
//     as file err (createFile "myFile.txt")
//     if (neq err "")
//         (println "Could not create file:" err)
//         return
//     as bytes (S<Byte> (Byte 100) (Byte 2) (Byte 255))
//     as n err (writeFile file bytes)                       // if no error, n should be same as length of bytes
//                                                           // the circumstances under which a write is partial are fairly exotic, but they can happen
//     if (neq err "")
//         (println "Could not write to file:" err)
//         return
//     as bytes (byteslice "hello, file world")
//     as n err (writeFile file bytes)
//     if (neq err "")
//         (println "Could not write to file:" err)
//         return
//     as err (closeFile file)
//     if (neq err "")
//         (println "Could not close file:" err)
//         return





func main
    locals file I err Str bytes S<Byte> n I msg Str
    as file err (openFile "myFile.txt")
    if (neq err "")
        (println "Could not open file:" err)
        return
    as bytes (make S<Byte> 1000)                    // logically doesn't matter how big our buffer is for this code, 
                                                    // but buffer size has performance consequences
    while true
        // even if no error, n may be less than length of bytes
        // the circumstances under which a read is partial are very common 
        // (storage drives can be relatively very slow, so rather than wait, 
        // the readFile operator will just read what it can and return) 
        as n err (readFile file bytes)              
        if (neq err "")
            if (eq err "EOF")
                break
            (println "Could not read from file:" err)
            return
        // use Str to convert a slice of bytes to a string
        as msg (concat msg (Str (slice bytes 0 n)))               
    // print out the whole file
    (print msg)       
    as err (closeFile file)
    if (neq err "")
        (println "Could not close file:" err)
        return
