import React, { useState, useEffect } from "react";
import { useParams } from "react-router-dom";
import GroupUsers from "./GroupUsers";
import Users from "./Users";
import Expense from "./Expense";
import Items from "./Items";
import Settlements from "./Settlements";

const Group = () => {
    const { groupId, groupName } = useParams();
    const [showUsersModal, setShowUsersModal] = useState(false);
    const [showExpenseModal, setShowExpenseModal] = useState(false);
    const [users, setUsers] = useState([]);
    const [loading, setLoading] = useState(true);
    const [refreshItems, setRefreshItems] = useState(0);

    const fetchGroupUsers = async () => {
        setLoading(true);
        try {
            const response = await fetch(`http://localhost:4000/api/groupUsers/${groupId}`);
            const data = await response.json();
            setUsers(data);
        } catch (error) {
            console.error("Error fetching group users:", error);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {    
        fetchGroupUsers();
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [groupId]);
    
    const handleOpenUsersModal = () => {
        setShowUsersModal(true);
    };

    const handleCloseUsersModal = () => {
        setShowUsersModal(false);
    };

    const handleOpenExpenseModal = () => {
        setShowExpenseModal(true);
    };

    const handleCloseExpenseModal = () => {
        setShowExpenseModal(false);
    };

    const handleUsersAdded = () => {
        // Close the modal and refresh group data
        setShowUsersModal(false);
        fetchGroupUsers(); // Refresh group data after adding users
    };

    const handleExpenseAdded = () => {
        // Close the expense modal after successful submission
        setShowExpenseModal(false);
        // Refresh the items list
        setRefreshItems(prev => prev + 1);
    };

    if (loading && users.length === 0) {
        return (
            <div className="flex justify-center items-center min-h-[40vh]">
                <div className="animate-pulse flex flex-col items-center">
                    <div className="h-8 w-64 bg-gray-200 rounded mb-4"></div>
                    <div className="h-4 w-32 bg-gray-200 rounded"></div>
                </div>
            </div>
        );
    }

    return (
        <div className="max-w-6xl mx-auto p-4">
            <div className="flex flex-col md:flex-row gap-6">
                {/* Main Content */}
                <div className="flex-grow">
                    <div className="bg-white shadow-md rounded-lg p-6 mb-6">
                        <div className="flex items-center justify-between mb-4">
                            <h1 className="text-2xl font-bold text-gray-800">
                                {groupName}
                            </h1>
                        </div>
                        
                        <div className="flex flex-wrap items-center gap-2 mb-6">
                            <button 
                                onClick={handleOpenExpenseModal} 
                                className="inline-flex items-center px-4 py-2 bg-green-600 hover:bg-green-700 text-white font-medium rounded-md shadow-sm transition-colors duration-200"
                            >
                                <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5 mr-2" viewBox="0 0 20 20" fill="currentColor">
                                    <path fillRule="evenodd" d="M4 4a2 2 0 00-2 2v4a2 2 0 002 2V6h10a2 2 0 00-2-2H4zm2 6a2 2 0 012-2h8a2 2 0 012 2v4a2 2 0 01-2 2H8a2 2 0 01-2-2v-4zm6 4a2 2 0 100-4 2 2 0 000 4z" clipRule="evenodd" />
                                </svg>
                                Add Expense
                            </button>
                        </div>
                        
                        {/* Settlements Component */}
                        <div className="border-t pt-4">
                            <Settlements groupId={groupId} users={users} refreshItems={refreshItems}/>
                        </div>
                    </div>
                    
                    <div className="bg-white shadow-md rounded-lg p-6">
                        <div className="flex justify-between items-center mb-6">
                            <h2 className="text-xl font-bold text-gray-800">Recent Expenses</h2>
                            {/* Optional: Add filter or sorting controls here */}
                        </div>
                        
                        {/* Items/Expenses list */}
                        <Items 
                            groupId={groupId} 
                            users={users} 
                            key={`items-${refreshItems}`} 
                        />
                    </div>
                </div>

                {/* Side Panel */}
                <div className="w-full md:w-80 flex-shrink-0">
                    <div className="bg-white shadow-md rounded-lg p-4 sticky top-4">
                        <div className="flex flex-col">
                            <GroupUsers users={users} />
                            
                            <button 
                                onClick={handleOpenUsersModal} 
                                className="mt-4 inline-flex items-center justify-center px-4 py-2 bg-indigo-600 hover:bg-indigo-700 text-white font-medium rounded-md shadow-sm transition-colors duration-200 w-full"
                            >
                                <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5 mr-2" viewBox="0 0 20 20" fill="currentColor">
                                    <path d="M8 9a3 3 0 100-6 3 3 0 000 6zM8 11a6 6 0 016 6H2a6 6 0 016-6zM16 7a1 1 0 10-2 0v1h-1a1 1 0 100 2h1v1a1 1 0 102 0v-1h1a1 1 0 100-2h-1V7z" />
                                </svg>
                                Add Members
                            </button>
                        </div>
                    </div>
                </div>
            </div>

            {/* Users Modal */}
            {showUsersModal && (
                <div className="fixed inset-0 z-50 flex items-center justify-center">
                    <div 
                        className="absolute inset-0 bg-black bg-opacity-60 backdrop-blur-sm transition-opacity duration-300"
                        onClick={handleCloseUsersModal}
                    ></div>
                    <div className="bg-white rounded-xl shadow-xl z-10 w-full max-w-md mx-4 relative transition-all duration-300 transform scale-100">
                        <div className="sticky top-0 bg-white p-4 border-b flex justify-between items-center rounded-t-xl">
                            <h2 className="text-xl font-bold text-gray-800">Add Members</h2>
                            <button 
                                onClick={handleCloseUsersModal}
                                className="p-1 rounded-full hover:bg-gray-100 transition-colors duration-200"
                            >
                                <svg xmlns="http://www.w3.org/2000/svg" className="h-6 w-6 text-gray-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                                </svg>
                            </button>
                        </div>
                        <div className="p-4">
                            <Users 
                                groupId={groupId}
                                inModal={true} 
                                onUsersAdded={handleUsersAdded}
                                onCancel={handleCloseUsersModal}
                            />
                        </div>
                    </div>
                </div>
            )}

            {/* Expense Modal */}
            {showExpenseModal && (
                <div className="fixed inset-0 z-50 flex items-center justify-center">
                    <div 
                        className="absolute inset-0 bg-black bg-opacity-60 backdrop-blur-sm transition-opacity duration-300"
                        onClick={handleCloseExpenseModal}
                    ></div>
                    <div className="bg-white rounded-xl shadow-xl z-10 w-full max-w-lg mx-4 relative transition-all duration-300 transform scale-100">
                        <div className="sticky top-0 bg-white p-4 border-b flex justify-between items-center rounded-t-xl">
                            <h2 className="text-xl font-bold text-gray-800">Add Expense</h2>
                            <button 
                                onClick={handleCloseExpenseModal}
                                className="p-1 rounded-full hover:bg-gray-100 transition-colors duration-200"
                            >
                                <svg xmlns="http://www.w3.org/2000/svg" className="h-6 w-6 text-gray-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                                </svg>
                            </button>
                        </div>
                        <div className="p-4">
                            <Expense 
                                users={users}
                                groupId={groupId}
                                onExpenseAdded={handleExpenseAdded}
                                onCancel={handleCloseExpenseModal}
                            />
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
};

export default Group;