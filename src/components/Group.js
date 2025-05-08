import React, { useState, useEffect } from "react";
import { useParams } from "react-router-dom";
import GroupUsers from "./GroupUsers";
import Users from "./Users";
import Expense from "./Expense";
import Items from "./Items";
import Settlements from "./Settlements";
import Memories from "./Memories"; // Import the new Memories component
import RecentTransactions from "./RecentTransactions";

const Group = () => {
    const { groupId, groupName } = useParams();
    const [showUsersModal, setShowUsersModal] = useState(false);
    const [showExpenseModal, setShowExpenseModal] = useState(false);
    const [showMemoriesModal, setShowMemoriesModal] = useState(false); // New state for Memories modal
    const [users, setUsers] = useState([]);
    const [loading, setLoading] = useState(true);
    const [refreshItems, setRefreshItems] = useState(0);
    const [refreshMemories, setRefreshMemories] = useState(0); // New state for refreshing memories
    const [refreshTransactions, setRefreshTransactions] = useState(0);

    const fetchGroupUsers = async () => {
        setLoading(true);
        try {
            const response = await fetch(`${process.env.REACT_APP_BACKEND_URL}/api/groupUsers/${groupId}`);
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

    // New handlers for Memories modal
    const handleOpenMemoriesModal = () => {
        setShowMemoriesModal(true);
    };

    const handleCloseMemoriesModal = () => {
        setShowMemoriesModal(false);
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

    // New handler for memory added
    const handleMemoryAdded = () => {
        // Refresh memories when new one is added
        setRefreshMemories(prev => prev + 1);
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

                            {/* Memories Button */}
                            <button 
                                onClick={handleOpenMemoriesModal} 
                                className="inline-flex items-center px-4 py-2 bg-purple-600 hover:bg-purple-700 text-white font-medium rounded-md shadow-sm transition-colors duration-200"
                            >
                                <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5 mr-2" viewBox="0 0 20 20" fill="currentColor">
                                    <path fillRule="evenodd" d="M4 3a2 2 0 00-2 2v10a2 2 0 002 2h12a2 2 0 002-2V5a2 2 0 00-2-2H4zm12 12H4l4-8 3 6 2-4 3 6z" clipRule="evenodd" />
                                </svg>
                                Memories
                            </button>
                        </div>
                        
                        {/* Settlements Component */}
                        <div className="border-t pt-4">
                            <Settlements groupId={groupId} users={users} refreshItems={refreshItems} setRefreshTransactions={setRefreshTransactions}/>
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

                <div className="w-full md:w-80 flex-shrink-0 flex flex-col gap-6">
                    <div className="sticky top-4 flex flex-col gap-6">
                        {/* Users section */}
                        <div className="bg-white shadow-md rounded-lg p-4">
                        <GroupUsers users={users} />

                        <button
                            onClick={handleOpenUsersModal}
                            className="mt-4 inline-flex items-center justify-center px-4 py-2 bg-indigo-600 hover:bg-indigo-700 text-white font-medium rounded-md shadow-sm transition-colors duration-200 w-full"
                        >
                            Add Members
                        </button>
                        </div>

                        {/* Recent Transactions component */}
                        <RecentTransactions users={users} refreshTransactions={refreshTransactions} />
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

        {/* Memories Modal */}
        {showMemoriesModal && (
            <div className="fixed inset-0 z-50">
                <div 
                    className="absolute inset-0 bg-black bg-opacity-70 backdrop-blur-sm transition-opacity duration-300"
                    onClick={handleCloseMemoriesModal}
                ></div>
                <div className="fixed inset-0 z-10 flex items-center justify-center pointer-events-none">
                    <div className="bg-white rounded-2xl shadow-2xl pointer-events-auto w-full max-w-3xl mx-4 relative transition-all duration-300 transform scale-100 flex flex-col max-h-[90vh] overflow-hidden">
                        {/* Enhanced header with gradient background */}
                        <div className="sticky top-0 z-10 bg-gradient-to-r from-indigo-600 to-purple-600 p-4 border-b flex justify-between items-center rounded-t-2xl">
                            <h2 className="text-xl font-bold text-white">Group Memories</h2>
                            <button 
                                onClick={handleCloseMemoriesModal}
                                className="p-1 rounded-full hover:bg-white hover:bg-opacity-20 transition-colors duration-200"
                            >
                                <svg xmlns="http://www.w3.org/2000/svg" className="h-6 w-6 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                                </svg>
                            </button>
                        </div>
                        
                        {/* Content area with memories component */}
                        <div className="flex-grow overflow-auto">
                            <Memories 
                                groupId={groupId}
                                onMemoryAdded={handleMemoryAdded}
                                key={`memories-${refreshMemories}`}
                            />
                        </div>
                    </div>
                </div>
            </div>
        )}
        </div>
    );
};

export default Group;