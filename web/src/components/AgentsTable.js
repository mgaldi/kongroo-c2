import React, { useState, useMemo, useEffect, useCallback } from 'react';
import useWebSocket from '../hooks/useWebSocket';
import Console from "./Console"
import { Table, Comment, Avatar, Card } from "antd";

function AgentsTable() {
  const [agentsData, setAgentsData] = useState(new Map())
  const [agents, setAgents] = useState(new Map())
  const { socket, socketConnected, connect, disconnect } = useWebSocket();
  const [agentsSet, setAgentsSet] = useState(new Set())
  const [dataTable, setDataTable] = useState([{}]);
  const [selectionType, setSelectionType] = useState('radio');
  const [selection, setSelection] = useState(false);
  const [selectedAgent, setSelectedAgent] = useState({});
  useEffect(() => {

    const fetchAgents = async () => {
      try {
        let res = await fetch('http://10.0.0.8:8080/agents/getallbase')
        res = await res.json()
        // setAgentsSet(prev => {
        //   let newSet = new Set()
        //   if (prev.length != 0) {
        //     newSet = new Set(prev)
        //   }
        //   for (let i = 0; i < res.length; i++) {
        //     newSet.add({ key: i, name: res[i].name, ip: res[i].ip, platform: res[i].platform })
        //   }
        //   return newSet
        // })
        setDataTable(prev => {
          const dArray = []
          for (let i = 0; i < res.length; i++) {
            let ob = { key: i, name: res[i].name, ip: res[i].ip, platform: res[i].platform, pcid : res[i].pcid}
            dArray.push(ob)
          }
          return dArray
        })
      } catch (error) {
        console.log(error);
      }
    }
    fetchAgents()
  }, [])

  //     setDataTable(prev => {
  //       let arr = agentsTwo.map((item, index) => {
  //         {
  //           key: index,
  //           name: item.name,
  //           ip: item.ip,
  //           platform: item.platform
  //         }
  //       })
  //     })
  // }
  const [inputValue, setInputValue] = useState('')
  // useEffect(() => {
  //   socket.onmessage = (e) => {
  //     console.log("Get message from server: " + e.data)
  //     let a = JSON.parse(e.data)
  //     //agentsData should be removed and instead just use agents.keys()

  //     setAgentsData(prevMap => {
  //       let newMap = new Map()
  //       if (prevMap.keys.length !== 0) {
  //         newMap = new Map(prevMap)
  //       }
  //       newMap.set(a.agent.Name, a.agent)

  //       return newMap
  //     })

  //     setAgentHistory(a.agent.Name)
  //   };

  //   return () => {
  //     disconnect
  //   }
  // }, [])
  useEffect(() => {
    socket.onmessage = (e) => {
      console.log("Get message from server: " + e.data)
      let a = JSON.parse(e.data)
      //agentsData should be removed and instead just use agents.keys()
      setAgents(prevMap => {
        let newMap = new Map()
        if (prevMap.size !== 0) {
          newMap = new Map(prevMap)
        }
        if(!newMap.has(a.agent.pcid)){
          
          retrieveAgentHistory(a.agent.pcid)
        } else {

          newMap.get(a.agent.pcid).push({command: a.agent.command, output: a.agent.output})
        }
        return newMap
      })
      // setAgentsData(prevMap => {
      //   let newMap = new Map()
      //   if (prevMap.keys.length !== 0) {
      //     newMap = new Map(prevMap)
      //   }
      //   newMap.set(a.agent.name, a.agent)

      //   return newMap
      // })
      // retrieveAgentHistory(a.agent.name)
    };

  }, [socket.onmessage])

  const extractAgents = () => {
    return dataTable.map(agents => agents.name);

  }

  const retrieveAgentHistory = (name) => {
    console.log("Request history for " + name)
    fetch(`http://10.0.0.8:8080/tasks/${name}/history`)
      .then(response => response.json())
      .then(data => {
        console.log("Retrieved History " + JSON.stringify(data))
        setAgents(prevMap => {
          console.log("SIZE " + prevMap.size + "PREV MAP " + Array.from(prevMap.keys()))
          let newMap = new Map()
          if (prevMap.size !== 0) {
            console.log("KEYS NOT ZERO")
            newMap = new Map(prevMap)
          }
          newMap.set(name, data)
          console.log(Array.from(newMap.keys()))
          return newMap
        })
      })

  }
  const handleClick = useCallback((e) => {
    e.preventDefault()

    socket.send(JSON.stringify({
      message: inputValue
    }))
  }, [inputValue])

  const agentRows = (arr) =>
    arr &&
    arr.map((item, index) => (
      <tr key={index} className="bg-gray-100">
        <td className="border px-4 py-2">{item.name}</td>
        <td className="border px-4 py-2">{item.ip}</td>
        <td className="border px-4 py-2">{item.platform}</td>
      </tr>
    ));
  const agentHead = (title) => (
    <thead>
      <tr>
        <th colSpan="4">{title}</th>
      </tr>
      <tr>
        <th className="px-4 py-2">Name</th>
        <th className="px-4 py-2">IP</th>
        <th className="px-4 py-2">Platform</th>
      </tr>
    </thead>
  )

  const tabNames = Array.from(agentsData.values()).map(x => {
    console.log(x)
    return <Console name={x.name} history={agents.get(x.name)}></Console>
  }
  )
  const [isPanel, setPanel] = useState(false)
  function showPanel() {
    setPanel(x => !x)
  }
  const rowSelection = {
    onChange: (selectedRowKeys, selectedRows) => {
      console.log(
        `selectedRowKeys: ${selectedRowKeys}`,
        "selectedRows: ",
        selectedRows
      );
    },
    getCheckboxProps: (record) => ({
      disabled: record.name === "Disabled User", // Column configuration not to be checked
      name: record.name,
    }),
  };
  const columns = [
    {
      title: "Agent",
      dataIndex: "name",
      key: "name",
      width: 200,
      // render: (text: string) => <a>{text}</a>,
    },
    {
      title: "IP",
      dataIndex: "ip",
      key: "ip",
      width: 200,
    },
    {
      title: "Platform",
      dataIndex: "platform",
      key: "platform",
      width: 200,
    },
  ];


  function showConsole(record) {
    console.log("From onRow - onClick" + JSON.stringify(record))
    retrieveAgentHistory(record.pcid)
    setSelectedAgent(() => ({
      pcid: record.pcid, name: record.name}))
  }
  function clickedRow(record) {
    showConsole(record)
    // setSelection(x => !x)
  }
  return (
    <>
      <div className="container">
        <h1 className="text-4xl text-red-600 text-center">KONGROO</h1>
        {/* {console.log("WE " + dataTable)} */}
        <Table
          onRow={(record) => {
            return {
              onClick: () => clickedRow(record)
            };
          }}
          rowSelection={{
            type: selectionType,
            ...rowSelection,
          }}
          columns={columns}
          dataSource={dataTable}
        />
      </div>
      <div className="mt-5 container">
        {<Console agent={selectedAgent} agents={agents}></Console>
          // (
          //   <div className="bg-slate-200 py-3">
          //     <span className="ml-5 font-bold">Agent Name </span>
          //     <Comment
          //       className="ml-5"
          //       avatar={
          //         <Avatar
          //           src="https://joeschmoe.io/api/v1/random"
          //           alt="Han Solo"
          //         />
          //       }
          //       content={
          //         <p>
          //           We supply a series of design principles, practical patterns
          //           and high quality design resources (Sketch and Axure), to help
          //           people create their product prototypes beautifully and
          //           efficiently.
          //           </p>
          //       }
          //     />
          //   </div>
          // )
        }
      </div>
    </>
    // <>
    //   <table className='table-auto'>
    //     {agentHead('Title')}
    //     <tbody>{agentRows(Array.from(agentsData.values()))}</tbody>
    //   </table>
    //   <section>
    //     {tabNames}
    //   </section>
    // </>
  )
}
export default AgentsTable;
