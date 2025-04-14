import React,{useEffect,useState} from "react";
import GroupCard from "./GroupCard";

const Groups = () =>{
    const [groups,setGroups] = useState([]);
    useEffect(()=>{
        const fetchGroups = async () => {
            const response = await fetch("http://localhost:4000/api/groupdetails");
            const data = await response.json();
            setGroups(data);
        };
        fetchGroups();
    },[]);

    return (
        <div>
            {
                groups.map((group)=>{
                    return <GroupCard key={group.group_id} group={group}/>;
                })
            }
        </div>
    )
}

export default Groups;