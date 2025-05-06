import React, { useState, useEffect, useMemo } from "react";
import { useAuth } from "../auth/AuthContext";

const Settlements = ({ groupId, users, refreshItems, setRefreshTransactions }) => {
  const { currentUser } = useAuth();
  const [settlements, setSettlements] = useState([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isOpen, setIsOpen] = useState(false);
  const [settlingUp, setSettlingUp] = useState(null);
  const [confirmSettlement, setConfirmSettlement] = useState(null);

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

  const handleSettleUpConfirm = (userId, amount) => {
    setConfirmSettlement({
      userId,
      amount,
      userName: findUserName(userId)
    });
  };

  const confirmAndSettleUp = async () => {
    if (!confirmSettlement) return;
    
    const { userId, amount } = confirmSettlement;
    setSettlingUp(userId);
    setConfirmSettlement(null);
    
    try {
      // Replace with your actual API endpoint for settling up
      const response = await fetch(`http://localhost:4000/api/insertTransactions/${groupId}`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          payer_id: currentUser.id,
          user_id: userId,
          amount: Math.abs(amount)
        }),
      });
      
      if (!response.ok) {
        throw new Error('Failed to settle up');
      }
      
      // Refresh settlements after successful settlement
      await fetchSettlements();
    } catch (error) {
      console.error("Error settling up:", error);
      // Optionally show an error message to the user
    } finally {
      setSettlingUp(null);
      setRefreshTransactions(prev => prev + 1);
    }
  };

  useEffect(() => {
    if (!users || !currentUser) return;
    fetchSettlements();
  }, [users, groupId, currentUser?.id, refreshItems]);

  // Check if there are any non-zero settlements
  const hasSettlements = useMemo(() => {
    if(!settlements)return false;
    return settlements.some(s => s.share_amount !== 0);
  }, [settlements]);

  const { totalBalance, isPositiveBalance } = useMemo(() => {
    if(!settlements)return false;
    const total = settlements.reduce((sum, s) => sum + s.share_amount, 0).toFixed(2);
    return {
      totalBalance: total,
      isPositiveBalance: parseFloat(total) >= 0
    };
  }, [settlements]);

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
                    className={`flex items-center justify-between p-3 rounded-lg text-sm ${
                      isPositive 
                        ? "bg-green-50 border border-green-100" 
                        : "bg-red-50 border border-red-100"
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
                          <span className="font-medium">{userName} owes you</span>
                        ) : (
                          <span className="font-medium">You owe {userName}</span>
                        )}
                      </div>
                    </div>
                    
                    <div className="flex items-center space-x-3">
                      {isPositive &&<div className={`font-medium ${isPositive ? "text-green-600" : "text-red-600"}`}>
                        ₹{amount}
                      </div>}
                      {!isPositive && (
                        <button
                          onClick={() => handleSettleUpConfirm(settlement.user_id, settlement.share_amount)}
                          disabled={settlingUp === settlement.user_id}
                          className={`px-3 py-1.5 rounded-full text-xs font-medium transition-all transform hover:scale-105 ${
                            "bg-red-600 hover:bg-red-700 text-white"
                          } ${settlingUp === settlement.user_id ? "opacity-70 cursor-wait" : ""}`}
                        >
                          {settlingUp === settlement.user_id ? (
                            <span className="flex items-center">
                              <svg className="animate-spin -ml-1 mr-2 h-3 w-3 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                                <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                                <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                              </svg>
                              Processing
                            </span>
                          ) : (
                            `Pay ₹${amount}`
                          )}
                        </button>
                      )}
                    </div>
                  </div>
                );
              })}
            </div>
          </div>
        )}

        {confirmSettlement && (
          <div className="fixed inset-0 z-50 flex items-center justify-center">
            <div 
              className="absolute inset-0 bg-black bg-opacity-50 backdrop-blur-sm"
              onClick={() => setConfirmSettlement(null)}
            ></div>
            <div className="bg-white rounded-lg shadow-xl z-10 w-full max-w-md mx-4 overflow-hidden">
              <div className="bg-indigo-50 p-4 border-b border-indigo-100">
                <h3 className="text-lg font-medium text-indigo-900">Confirm Settlement</h3>
              </div>
              <div className="p-5">
                <div className="flex items-center justify-center mb-5">
                <div className="h-16 w-16 bg-indigo-100 rounded-full flex items-center justify-center">
                  <span className="text-2xl font-bold text-indigo-600">₹</span>
                </div>
                </div>
                
                <p className="text-center text-gray-700 mb-6">
                  You're about to settle up <span className="font-medium text-gray-900">₹{Math.abs(confirmSettlement.amount).toFixed(2)}</span> with <span className="font-medium text-gray-900">{confirmSettlement.userName}</span>.
                </p>
                
                <div className="flex space-x-3">
                  <button 
                    onClick={() => setConfirmSettlement(null)}
                    className="flex-1 py-2.5 bg-gray-100 hover:bg-gray-200 rounded-lg text-gray-800 font-medium transition-colors"
                  >
                    Cancel
                  </button>
                  <button 
                    onClick={confirmAndSettleUp}
                    className="flex-1 py-2.5 bg-indigo-600 hover:bg-indigo-700 rounded-lg text-white font-medium transition-colors"
                    disabled={settlingUp === confirmSettlement.userId}
                  >
                    {settlingUp === confirmSettlement.userId ? (
                      <span className="flex items-center justify-center">
                        <svg className="animate-spin -ml-1 mr-2 h-4 w-4 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                          <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                          <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                        </svg>
                        Processing...
                      </span>
                    ) : (
                      "Confirm Settlement"
                    )}
                  </button>
                </div>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

export default Settlements;