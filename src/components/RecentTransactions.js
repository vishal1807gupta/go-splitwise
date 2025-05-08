import React from "react"; 
import { useState, useEffect } from "react"; 
import { useParams } from "react-router-dom"; 
import { useAuth } from "../auth/AuthContext";  

const RecentTransactions = ({users, refreshTransactions}) => {
    const params = useParams();
    const groupId = params.groupId;
    const [transactions, setTransactions] = useState([]);
    const [loading, setLoading] = useState(true);
    const { currentUser } = useAuth();
    
    const fetchTransactions = async () => {
        setLoading(true);
        try {
            const response = await fetch(`${process.env.REACT_APP_BACKEND_URL}/api/getTransactions/${groupId}`);
            const data = await response.json();
            setTransactions(data);
        } catch(error) {
            console.error("Error fetching transactions:", error);
        } finally {
            setLoading(false);
        }
    }
    
    const findUserName = (userId) => {
        const user = users.find(u => u.id === userId);
        return user ? user.name : `User ${userId}`;
    };
    
    useEffect(() => {
        fetchTransactions();
    }, [groupId, refreshTransactions]);
    
    // Format date to be more readable
    const formatDate = (dateString) => {
        const date = new Date(dateString);
        const now = new Date();
        
        // Check if date is today
        if (date.toDateString() === now.toDateString()) {
            return 'Today';
        }
        
        // Check if date was yesterday
        const yesterday = new Date(now);
        yesterday.setDate(now.getDate() - 1);
        if (date.toDateString() === yesterday.toDateString()) {
            return 'Yesterday';
        }
        
        // Return formatted date
        return date.toLocaleDateString([], {
            year: 'numeric',
            month: 'short',
            day: 'numeric'
        });
    };
    
    // Get transaction icon based on user's involvement
    const getTransactionIcon = (transaction) => {
        if (transaction.payer_id === currentUser.id) {
            // Current user paid
            return (
                <div className="w-10 h-10 rounded-full bg-red-100 flex items-center justify-center flex-shrink-0">
                    <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5 text-red-600" viewBox="0 0 20 20" fill="currentColor">
                        <path fillRule="evenodd" d="M14.707 12.707a1 1 0 01-1.414 0L10 9.414l-3.293 3.293a1 1 0 01-1.414-1.414l4-4a1 1 0 011.414 0l4 4a1 1 0 010 1.414z" clipRule="evenodd" />
                    </svg>
                </div>
            );
        } else if (transaction.user_id === currentUser.id) {
            // Current user received payment
            return (
                <div className="w-10 h-10 rounded-full bg-green-100 flex items-center justify-center flex-shrink-0">
                    <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5 text-green-600" viewBox="0 0 20 20" fill="currentColor">
                        <path fillRule="evenodd" d="M5.293 7.293a1 1 0 011.414 0L10 10.586l3.293-3.293a1 1 0 111.414 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 010-1.414z" clipRule="evenodd" />
                    </svg>
                </div>
            );
        } else {
            // Transaction between other users
            return (
                <div className="w-10 h-10 rounded-full bg-gray-100 flex items-center justify-center flex-shrink-0">
                    <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5 text-gray-600" viewBox="0 0 20 20" fill="currentColor">
                        <path fillRule="evenodd" d="M4 2a1 1 0 011 1v2.101a7.002 7.002 0 0111.601 2.566 1 1 0 11-1.885.666A5.002 5.002 0 005.999 7H9a1 1 0 010 2H4a1 1 0 01-1-1V3a1 1 0 011-1zm.008 9.057a1 1 0 011.276.61A5.002 5.002 0 0014.001 13H11a1 1 0 110-2h5a1 1 0 011 1v5a1 1 0 11-2 0v-2.101a7.002 7.002 0 01-11.601-2.566 1 1 0 01.61-1.276z" clipRule="evenodd" />
                    </svg>
                </div>
            );
        }
    };

    if (loading) {
        return (
            <div className="bg-white shadow-md rounded-lg p-4 mt-6">
                <h2 className="text-lg font-bold text-gray-800 mb-4">Recent Transactions</h2>
                <div className="space-y-3">
                    {[...Array(3)].map((_, i) => (
                        <div key={i} className="flex items-center animate-pulse">
                            <div className="w-10 h-10 bg-gray-200 rounded-full mr-3"></div>
                            <div className="flex-grow">
                                <div className="h-4 bg-gray-200 rounded w-3/4 mb-2"></div>
                                <div className="h-3 bg-gray-200 rounded w-1/2"></div>
                            </div>
                            <div className="h-4 bg-gray-200 rounded w-16"></div>
                        </div>
                    ))}
                </div>
            </div>
        );
    }

    return (
        <div className="bg-white shadow-md rounded-lg p-4 mt-6">
            <h2 className="text-lg font-bold text-gray-800 mb-4">Recent Transactions</h2>
            
            {!transactions || transactions.length === 0 ? (
                <div className="text-center py-8 border border-dashed border-gray-200 rounded-lg">
                    <svg xmlns="http://www.w3.org/2000/svg" className="h-12 w-12 mx-auto text-gray-400 mb-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
                    </svg>
                    <p className="text-gray-500">No transactions yet</p>
                </div>
            ) : (
                <div className="space-y-3 max-h-72 overflow-y-auto pr-1 scrollbar-thin scrollbar-thumb-gray-300 scrollbar-track-gray-100">
                    {transactions && transactions.length>0 && transactions.map((transaction) => {
                        const isUserPayer = transaction.payer_id === currentUser.id;
                        const isUserReceiver = transaction.user_id === currentUser.id;
                        
                        return (
                            <div key={transaction.id} className="bg-white border border-gray-100 rounded-lg p-2 hover:bg-gray-50 transition-colors">
                                <div className="flex items-center space-x-3">
                                    {getTransactionIcon(transaction)}
                                    
                                    <div className="flex-grow">
                                        <div className="flex items-center justify-between">
                                            <div className="text-xs text-gray-700">
                                                {isUserPayer ? (
                                                    <>You paid <span className="text-red-600 font-medium">₹{transaction.amount}</span> to {findUserName(transaction.user_id)}</>
                                                ) : isUserReceiver ? (
                                                    <><span className="text-green-600 font-medium">₹{transaction.amount}</span> received from {findUserName(transaction.payer_id)}</>
                                                ) : (
                                                    <>{findUserName(transaction.payer_id)} paid <span className="text-indigo-600 font-medium">₹{transaction.amount}</span> to {findUserName(transaction.user_id)}</>
                                                )}
                                            </div>
                                            <div className="text-xs text-gray-400 ml-2">
                                                {formatDate(transaction.created_at)}
                                            </div>
                                        </div>
                                    </div>
                                </div>
                            </div>
                        );
                    })}
                </div>
            )}
        </div>
    );
};

export default RecentTransactions;