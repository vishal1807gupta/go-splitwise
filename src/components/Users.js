import React from "react";
import { useEffect, useState } from "react";
import { useParams, useNavigate } from "react-router-dom";
import User from "./User";

const Users = ({ groupId: propGroupId, inModal = false, onUsersAdded, onCancel }) => {
    const navigate = useNavigate();
    // Use the groupId from props if provided (modal mode), otherwise from URL params
    const { groupId: urlGroupId } = useParams();
    const groupId = propGroupId || urlGroupId;
    
    const [users, setUsers] = useState([]);
    const [checkedUsers, setCheckedUsers] = useState([]);
    const [loading, setLoading] = useState(true);

    const toggle = (userId) => {
        setCheckedUsers((prevCheckedUsers) => {
            if (prevCheckedUsers.includes(userId)) {
                return prevCheckedUsers.filter((id) => id !== userId);
            } else {
                return [...prevCheckedUsers, userId];
            }
        });
    };

    const AddUsersToGroup = async () => {
        try {
            const response = await fetch(`http://localhost:4000/api/addUsersToGroup/${groupId}`, {
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                },
                body: JSON.stringify({ user_ids: checkedUsers }),
            });
            const data = await response.json();
            console.log(data);
            
            // If in modal, use callback instead of navigation
            if (inModal && onUsersAdded) {
                onUsersAdded();
            } else {
                navigate(`/group/${groupId}`);
            }
        } catch (error) {
            console.error("Error adding users to group:", error);
        }
    };

    useEffect(() => {
        setLoading(true);
        const fetchGroupUsers = async () => {
            try {
                const response = await fetch(`http://localhost:4000/api/notGroupUsers/${groupId}`);
                const data = await response.json();
                setUsers(data);
            } catch (error) {
                console.error("Error fetching users:", error);
            } finally {
                setLoading(false);
            }
        };
        fetchGroupUsers();
    }, [groupId]);

    if (loading) {
        return (
            <div className="w-full">
                <div className="space-y-2 animate-pulse">
                    {[...Array(3)].map((_, index) => (
                        <div key={index} className="flex items-center p-2 border rounded">
                            <div className="w-8 h-8 bg-gray-200 rounded-full mr-2"></div>
                            <div className="flex-grow">
                                <div className="h-3 bg-gray-200 rounded w-1/3 mb-1"></div>
                                <div className="h-2 bg-gray-200 rounded w-1/2"></div>
                            </div>
                            <div className="w-4 h-4 bg-gray-200 rounded ml-2"></div>
                        </div>
                    ))}
                </div>
            </div>
        );
    }

    return (
        <div className="w-full">
            {users && users.length > 0 ? (
                <>
                    <div className="space-y-1 mb-4 max-h-64 overflow-y-auto pr-1">
                        {users.map((user) => (
                            <div key={user.id} className="flex items-center p-2 border rounded hover:bg-gray-50">
                                <User user={user} compact={true} />
                                <input 
                                    type="checkbox"
                                    checked={checkedUsers.includes(user.id)}
                                    onChange={() => toggle(user.id)}
                                    className="ml-auto h-4 w-4 text-indigo-600 rounded"
                                />
                            </div>
                        ))}
                    </div>
                    
                    <div className="flex items-center justify-between border-t pt-3">
                        <span className="text-xs text-gray-500">
                            {checkedUsers.length} {checkedUsers.length === 1 ? 'user' : 'users'} selected
                        </span>
                        <div className="flex space-x-2">
                            {inModal && (
                                <button
                                    onClick={onCancel}
                                    className="px-3 py-1.5 border border-gray-300 rounded-md text-xs font-medium text-gray-700 bg-white hover:bg-gray-50"
                                >
                                    Cancel
                                </button>
                            )}
                            <button
                                onClick={AddUsersToGroup}
                                className="bg-indigo-600 hover:bg-indigo-700 text-white font-medium text-xs py-1.5 px-3 rounded"
                                disabled={checkedUsers.length === 0}
                            >
                                Add to Group
                            </button>
                        </div>
                    </div>
                </>
            ) : (
                <div className="text-center py-8 bg-gray-50 rounded-lg">
                    <svg xmlns="http://www.w3.org/2000/svg" className="h-8 w-8 mx-auto text-gray-400 mb-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M18 9v3m0 0v3m0-3h3m-3 0h-3m-2-5a4 4 0 11-8 0 4 4 0 018 0zM3 20a6 6 0 0112 0v1H3v-1z" />
                    </svg>
                    <p className="text-sm text-gray-500 mb-2">No users available to add</p>
                </div>
            )}
        </div>
    );
};

export default Users;