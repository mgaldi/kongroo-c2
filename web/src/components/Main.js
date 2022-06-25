import React, { useEffect, useState } from 'react'
import useWebSocket from '../hooks/useWebSocket';
import AgentsTable from './AgentsTable';
import MenuItem from './MenuItem';
import { StatusOfflineIcon, StatusOnlineIcon } from '@heroicons/react/solid'

function Main() {
    const { socket, socketConnected, connect } = useWebSocket();
    // const [agents, setAgents] = useState([]);

    // useEffect(() => async () => {
    //     fetchAgents()
    // }, [socket])


    // const fetchAgents = async () => {
    //     let res = await fetch('http://localhost:8080/agents/getall')
    //     res = await res.json()
    //     setAgents(res.agents)
    // }

    return (
        <div className="h-screen">
            {socketConnected ? (
                <div className="mx-44">
                    {/* <div className="w-2/12 h-screen border-r-2 border-[#fb4467]">
                        <p className="mt-8 text-center">
                            Agents List
                        </p>
                        <div className="flex ml-2 sm:ml-4 items-center my-4">
                            <input type="search"
                                className="
                                w-full
                                mx-8
                                px-3
                                py-1.5
                                bg-transparent
                                border border-solid border-gray-300
                                rounded
                                focus:border-[#fb4467] focus:outline-none"
                                id="searchAgent"
                                placeholder="Search Agent"
                            />

                        </div>
                        {agents.map((data, index) => (
                            < MenuItem
                                hostname={data} Icon={index % 2 === 0 ? StatusOnlineIcon : StatusOfflineIcon} key={index} />
                        ))}
                    </div> */}
                    <div className="">
                        <AgentsTable />
                    </div>
                </div>
            ) : (
                <div className="flex flex-col h-screen justify-center text-center text-3xl">
                    Web Socket Not Connected
                    <button className="text-red-400" onClick={connect}>Retry</button>
                </div>
            )}
        </div>
    )
}

export default Main
