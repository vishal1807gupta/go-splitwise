import React, { useState, useEffect } from "react";
import { useParams } from "react-router-dom";

const Group = () => {
    const { groupId } = useParams();
    const [group, setGroup] = useState(null);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        const fetchGroup = async () => {
            try {
                const response = await fetch(`http://localhost:4000/api/groupdetails/${groupId}`);
                const data = await response.json();
                setGroup(data);
            } catch (error) {
                console.error("Error fetching group:", error);
            } finally {
                setLoading(false);
            }
        };
        
        fetchGroup();
    }, [groupId]);

    if (loading) {
        return <div>Loading...</div>;
    }

    if (!group) {
        return <div>Group not found</div>;
    }

    return (
        <div>
            <h1>Hello Peeps! I'm {group.group_name} Group</h1>
        </div>
    );
};

export default Group;