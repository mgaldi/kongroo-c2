import React, { useState, useMemo, useEffect, useCallback } from 'react';
import { Buffer } from 'buffer';

export default function Console(props) {
    const { agent, agents } = props
    const [input, setInput] = useState("")
    function handleInputChange(event) {
        setInput(x => event.target.value)
    }
    function debugShowHistory() {
        console.log(agents.get(agent.pcid))
    }

    function handleSubmit(e) {
        e.preventDefault()
        let command = Buffer.from(input).toString('base64')
        console.log(command)
        try {
            fetch(`http://localhost:8080/tasks/${agent.pcid}/${command}`, {
                method: "POST"
            })
        } catch (err) {
            console.log(err)
        }
    }
    return (
        <div className="bg-slate-100 ">
            <div className="overflow-auto
                    p-4 mx-16 h-96 scroll-smooth">
{/* {console.log("Console for agent: " + name + "History: " + agents.get(name))} */}
            {agents.get(agent.pcid) && 
                (agents.get(agent.pcid)).map(x => {
                    return (
                        <pre>
                            <p>Command: {x.command}</p>
                            <p>Output: {x.output}</p>
                            <br></br>
                        </pre>

                    )
                })
            }</div>
            <form className="w-full max-w-sm" onSubmit={handleSubmit}>
                <div className="flex items-center border-b border-teal-500 py-2">
                    {/* <label className="flex-auto">
                        {props.name} */}
                        <input value={input} onChange={handleInputChange} className="appearance-none bg-transparent border-none w-full text-gray-700 mr-3 py-1 px-2 leading-tight focus:outline-none" type="text" placeholder={agent.name} aria-label="Full name" />
                    {/* </label> */}
                    <button type="submit" value="Exec" className="flex-shrink-0 bg-teal-500 hover:bg-teal-700 border-teal-500 hover:border-teal-700 text-sm border-4 text-white py-1 px-2 rounded" >
                        Exec
                    </button>
                </div>
            </form>
            {/* <form style={{ marginTop: 20 }} onSubmit={handleSubmit} className="flex" >
                <label className="flex-auto">
                    {props.name}
                    <input type="text" value={input} onChange={handleInputChange} className="mx-3" ></input>
                </label>
                <input type="submit" value="Exec" className="flex-auto" />

            </form> */}

        </div>
    )
}
