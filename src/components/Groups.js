import React, { useEffect, useState } from "react";
import GroupCard from "./GroupCard";
import GroupForm from "./GroupForm";
import {useAuth} from "../auth/AuthContext";

const Groups = () => {
    const [groups, setGroups] = useState([]);
    const [showModal, setShowModal] = useState(false);
    const [loading, setLoading] = useState(true);
    const [searchQuery, setSearchQuery] = useState("");
    const {currentUser} = useAuth();
    const userId = currentUser ? currentUser.id : null;

    const fetchGroups = async () => {
        setLoading(true);
        try {
            const response = await fetch(`http://localhost:4000/api/groupdetails/${userId}`);
            const data = await response.json();
            setGroups(data);
        } catch (error) {
            console.error("Error fetching groups:", error);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        if(currentUser)fetchGroups();
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [currentUser]);

    const handleCreateGroup = () => {
        setShowModal(true);
    };

    const handleCloseModal = () => {
        setShowModal(false);
    };

    const handleGroupCreated = () => {
        // Refresh the groups list
        fetchGroups();
        // Close the modal
        setShowModal(false);
    };
    
    // Filter groups based on search query
    const filteredGroups = groups && groups.filter(group => {
        const query = searchQuery.toLowerCase().trim();
        if (!query) return true;
        
        return group.group_name && group.group_name.toLowerCase().includes(query);
    });

    return (
        <div className="max-w-6xl mx-auto p-4">
            <div className="flex justify-between items-center mb-6">
                <h1 className="text-2xl font-bold text-gray-800">My Groups</h1>
                <button 
                    onClick={handleCreateGroup} 
                    className="inline-flex items-center px-4 py-2 bg-pink-600 hover:bg-pink-700 text-white font-medium rounded-md shadow-sm transition-colors duration-200"
                >
                    <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5 mr-2" viewBox="0 0 20 20" fill="currentColor">
                        <path fillRule="evenodd" d="M10 3a1 1 0 011 1v5h5a1 1 0 110 2h-5v5a1 1 0 11-2 0v-5H4a1 1 0 110-2h5V4a1 1 0 011-1z" clipRule="evenodd" />
                    </svg>
                    Create Group
                </button>
            </div>

            {/* Search bar */}
            {!loading && groups && groups.length > 0 && (
                <div className="mb-6">
                    <div className="relative">
                        <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                            <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
                            </svg>
                        </div>
                        <input
                            type="text"
                            className="block w-full pl-10 pr-3 py-2 border border-gray-300 rounded-md text-sm placeholder-gray-400 focus:outline-none focus:ring-pink-500 focus:border-pink-500"
                            placeholder="Search groups by name..."
                            value={searchQuery}
                            onChange={(e) => setSearchQuery(e.target.value)}
                        />
                        {searchQuery && (
                            <button
                                className="absolute inset-y-0 right-0 pr-3 flex items-center text-gray-400 hover:text-gray-600"
                                onClick={() => setSearchQuery("")}
                            >
                                <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                                </svg>
                            </button>
                        )}
                    </div>
                </div>
            )}

            {loading ? (
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                    {[...Array(6)].map((_, index) => (
                        <div key={index} className="animate-pulse bg-white shadow-md rounded-lg p-4 h-64">
                            <div className="h-6 bg-gray-200 rounded w-3/4 mb-4"></div>
                            <div className="h-4 bg-gray-200 rounded w-full mb-2"></div>
                            <div className="h-4 bg-gray-200 rounded w-full mb-2"></div>
                            <div className="h-4 bg-gray-200 rounded w-2/3"></div>
                            <div className="h-5 bg-gray-200 rounded w-1/3 mt-auto"></div>
                        </div>
                    ))}
                </div>
            ) : groups && groups.length > 0 ? (
                <>
                    {filteredGroups.length > 0 ? (
                        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                            {filteredGroups.map((group) => (
                                <div key={group.group_id} className="h-full">
                                    <GroupCard group={group} />
                                </div>
                            ))}
                        </div>
                    ) : (
                        <div className="text-center py-12 bg-gray-50 rounded-lg">
                            <svg xmlns="http://www.w3.org/2000/svg" className="h-12 w-12 mx-auto text-gray-400 mb-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
                            </svg>
                            <h3 className="text-lg font-medium text-gray-700 mb-2">No Groups Found</h3>
                            <p className="text-gray-500 mb-4">No groups match your search for "{searchQuery}"</p>
                            <button 
                                onClick={() => setSearchQuery("")} 
                                className="inline-flex items-center px-4 py-2 bg-gray-200 hover:bg-gray-300 text-gray-800 font-medium rounded-md"
                            >
                                Clear Search
                            </button>
                        </div>
                    )}
                </>
            ) : (
                <div className="text-center py-12 bg-gray-50 rounded-lg">
                    <svg xmlns="http://www.w3.org/2000/svg" className="h-12 w-12 mx-auto text-gray-400 mb-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
                    </svg>
                    <h3 className="text-lg font-medium text-gray-700 mb-2">No Groups Found</h3>
                    <p className="text-gray-500 mb-4">You haven't created or joined any groups yet.</p>
                    <button 
                        onClick={handleCreateGroup} 
                        className="inline-flex items-center px-4 py-2 bg-pink-600 hover:bg-pink-700 text-white font-medium rounded-md"
                    >
                        Create Your First Group
                    </button>
                </div>
            )}

            {/* Modal overlay with blur effect */}
            {showModal && (
                <div className="fixed inset-0 z-50 flex items-center justify-center">
                    <div 
                        className="absolute inset-0 bg-black bg-opacity-60 backdrop-blur-sm transition-opacity duration-300"
                        onClick={handleCloseModal}
                    ></div>
                    <div className="bg-white rounded-xl shadow-xl z-10 max-w-md w-full mx-4 relative transition-all duration-300 transform scale-100">
                        <div className="sticky top-0 bg-white p-4 border-b flex justify-between items-center rounded-t-xl">
                            <h2 className="text-xl font-bold text-gray-800">Create New Group</h2>
                            <button 
                                onClick={handleCloseModal}
                                className="p-1 rounded-full hover:bg-gray-100 transition-colors duration-200"
                            >
                                <svg xmlns="http://www.w3.org/2000/svg" className="h-6 w-6 text-gray-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                                </svg>
                            </button>
                        </div>
                        <div className="p-6">
                            <GroupForm onSuccess={handleGroupCreated} onClose={handleCloseModal} />
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
};

export default Groups;