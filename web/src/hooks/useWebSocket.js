import React, { createContext, useState, useMemo, useContext, useEffect } from 'react'

const SocketContext = createContext({});

export const SocketProvider = ({ children }) => {
    const [socketConnected, setSocketConnected] = useState(false);
    const [socket, setSocket] = useState(null);

    useEffect(() => {
        connect()
    }, [])

    const onMessage = (e) => {
        let a = JSON.parse(e.data)
        console.log(a)
    }

    const connect = () => {
        const ws = new WebSocket("ws://10.0.0.8:8080/ws")
        // Connection opened
        ws.addEventListener('open', function () {
            console.log('Connected WS')
            ws.addEventListener('message', event => onMessage(event));
            setSocket(ws)
            setSocketConnected(true)
        });


    }

    const disconnect = async () => {
        socket.close()
        setSocket(null)
    }



    const memoedValue = useMemo(() => ({
        socket,
        socketConnected,
        connect,
        disconnect,
    }), [socket, socketConnected])

    return (
        <SocketContext.Provider value={memoedValue}>
            {children}
        </SocketContext.Provider>
    )
}
export default function useWebSocket() {
    return useContext(SocketContext);
}
