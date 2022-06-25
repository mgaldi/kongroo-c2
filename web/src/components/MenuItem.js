import React from 'react'

function MenuItem({ hostname, Icon }) {
    return (
        < div className="flex flex-row items-center cursor-pointer group mt-2" >
            <div className="flex flex-row w-full justify-between mx-8 ">
                <p className="group-hover:animate-pulse group-hover:text-[#fb4467]">
                    {hostname}
                </p>
                <Icon width="20px" ml="12px" className="text-green-400" />
            </div>
        </div >

    )
}

export default MenuItem
