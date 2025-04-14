import React from "react";
import {useNavigate} from "react-router-dom";

const GroupCard = ({group}) =>{
    const navigate = useNavigate();
    return (
        <>
        <div>
          {group.group_name}
        </div>
        <button onClick={()=>{navigate(`/group/${group.group_id}`)}} className="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded">Open Group</button>
        </>
    )
}

export default GroupCard;