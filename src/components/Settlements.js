import React, { useState, useEffect } from "react";
import { useAuth } from "../auth/AuthContext";

const Settlements = ({ groupId, users }) => {
  const { currentUser } = useAuth();
  const [settlements, setSettlements] = useState([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isOpen, setIsOpen] = useState(false);

  const findUserName = (userId) => {
    const user = users.find(u => u.id === userId);
    return user ? user.name : `User ${userId}`;
  };

  const fetchSettlements = async () => {
    setIsLoading(true);
    try {
      const response = await fetch(
        `http://localhost:4000/api/settlements/${groupId}/${currentUser.id}`,
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({ users: users.map(user => user.id) }),
        }
      );
      const data = await response.json();
      setSettlements(data);
      console.log(data);
    } catch (error) {
      console.error("Error fetching settlements:", error);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    if (!users || !currentUser) return;
    fetchSettlements();
  }, [users, groupId, currentUser?.id]);

  // Check if there are any non-zero settlements
  const hasSettlements = settlements.some(s => s.share_amount !== 0);
  
  // Get total balance (positive amount means others owe you, negative means you owe others)
  const totalBalance = settlements.reduce((sum, s) => sum + s.share_amount, 0).toFixed(2);
  const isPositiveBalance = parseFloat(totalBalance) >= 0;

  if (isLoading) {
    return (
      <div className="mt-4 space-y-2 animate-pulse">
        <div className="h-6 bg-gray-200 rounded w-3/4"></div>
      </div>
    );
  }

  if (!hasSettlements) {
    return (
      <div className="mt-4 p-3 bg-gray-50 rounded-lg border border-gray-100 text-sm">
        <p className="text-center text-gray-500">All balances are settled in this group</p>
      </div>
    );
  }

  return (
    <div className="mt-4">
      <div 
        className={`border rounded-lg overflow-hidden transition-all duration-200 ${
          isOpen ? 'border-gray-300' : 'border-gray-200 hover:border-gray-300'
        }`}
      >
        {/* Accordion Header */}
        <button
          onClick={() => setIsOpen(!isOpen)}
          className="w-full px-4 py-3 flex items-center justify-between bg-gray-50 focus:outline-none"
        >
          <div className="flex items-center">
            <span className="font-medium text-gray-800">Balance Summary</span>
            <span className={`ml-3 px-2 py-1 text-xs font-medium rounded-full ${
              isPositiveBalance ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'
            }`}>
              {isPositiveBalance ? 'You are owed' : 'You owe'} ₹{Math.abs(totalBalance)}
            </span>
          </div>
          <svg
            className={`w-5 h-5 text-gray-500 transition-transform ${isOpen ? 'transform rotate-180' : ''}`}
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
          >
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
          </svg>
        </button>

        {/* Accordion Content */}
        {isOpen && (
          <div className="bg-white p-3 max-h-40 overflow-y-auto">
            <div className="space-y-2">
              {settlements.map((settlement) => {
                if (settlement.share_amount === 0) return null;
                
                const isPositive = settlement.share_amount > 0;
                const amount = Math.abs(settlement.share_amount).toFixed(2);
                const userName = findUserName(settlement.user_id);
                
                return (
                  <div 
                    key={settlement.user_id} 
                    className={`flex items-center justify-between p-2 rounded text-sm ${
                      isPositive 
                        ? "bg-green-50" 
                        : "bg-red-50"
                    }`}
                  >
                    <div className="flex items-center">
                      <div className={`mr-2 ${isPositive ? "text-green-600" : "text-red-600"}`}>
                        {isPositive ? (
                          <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4" viewBox="0 0 20 20" fill="currentColor">
                            <path fillRule="evenodd" d="M5.293 9.707a1 1 0 010-1.414l4-4a1 1 0 011.414 0l4 4a1 1 0 01-1.414 1.414L11 7.414V15a1 1 0 11-2 0V7.414L6.707 9.707a1 1 0 01-1.414 0z" clipRule="evenodd" />
                          </svg>
                        ) : (
                          <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4" viewBox="0 0 20 20" fill="currentColor">
                            <path fillRule="evenodd" d="M14.707 10.293a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 111.414-1.414L9 12.586V5a1 1 0 012 0v7.586l2.293-2.293a1 1 0 011.414 0z" clipRule="evenodd" />
                          </svg>
                        )}
                      </div>
                      <div>
                        {isPositive ? (
                          <span>{userName} owes you</span>
                        ) : (
                          <span>You owe {userName}</span>
                        )}
                      </div>
                    </div>
                    <div className={`font-medium ${isPositive ? "text-green-600" : "text-red-600"}`}>
                      ₹{amount}
                    </div>
                  </div>
                );
              })}
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

export default Settlements;