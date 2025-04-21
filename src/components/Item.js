import React, { useState } from "react";

const Item = ({ item, users }) => {
  const [isExpanded, setIsExpanded] = useState(false);
  
  if (!item || !item.user_shares) return null;

  // Find payer name if users array is provided
  const findUserName = (userId) => {
    const user = users.find(u => u.id === userId);
    return user ? user.name : `User ${userId}`;
  };
  
  return (
    <div className={`bg-white rounded-lg shadow-sm border-l-4 border-gray-400 overflow-hidden transition-all duration-200`}>
      {/* Header - Always visible */}
      <div
        className="p-3 cursor-pointer hover:bg-gray-50"
        onClick={() => setIsExpanded(!isExpanded)}
      >
        <div className="flex justify-between items-center">
          <div className="flex-grow">
            <h3 className="text-base font-medium text-gray-900">{item.description}</h3>
            <div className="flex items-center text-xs text-gray-600 mt-0.5">
              <svg xmlns="http://www.w3.org/2000/svg" className="h-3 w-3 mr-1 text-gray-500" viewBox="0 0 20 20" fill="currentColor">
                <path fillRule="evenodd" d="M10 9a3 3 0 100-6 3 3 0 000 6zm-7 9a7 7 0 1114 0H3z" clipRule="evenodd" />
              </svg>
              <span>Paid by: <span className="font-medium">{findUserName(item.payer_id)}</span></span>
            </div>
          </div>
          
          <div className="flex items-center">
            <div className="text-base font-bold text-gray-800 mr-2">₹{Math.abs(item.amount).toFixed(2)}</div>
            <svg
              xmlns="http://www.w3.org/2000/svg"
              className={`h-4 w-4 text-gray-400 transition-transform ${isExpanded ? 'transform rotate-180' : ''}`}
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
            </svg>
          </div>
        </div>
      </div>

      {/* Expandable content */}
      <div
        className={`border-t border-gray-100 overflow-hidden transition-all duration-300 ${
          isExpanded ? 'max-h-32 opacity-100' : 'max-h-0 opacity-0'
        }`}
      >
        <div className="p-2">
          <div className="text-xs text-gray-500 mb-1">Split Details:</div>
          <div className="space-y-1 overflow-y-auto max-h-20">
            {item.user_shares.map((share, index) => (
              <div key={index} className="flex justify-between items-center py-1 px-2 rounded bg-gray-50 text-xs">
                <div>{findUserName(share.user_id)}</div>
                <div className={`font-medium ${share.share_amount > 0 ? 'text-green-600' : 'text-red-500'}`}>
                  {share.share_amount > 0 ? '+' : ''}{share.share_amount.toFixed(2)}
                </div>
              </div>
            ))}
          </div>
          
          {/* Date if available */}
          {item.date && (
            <div className="text-xs text-gray-400 mt-1 text-right">
              {new Date(item.date).toLocaleString([], {
                year: 'numeric',
                month: 'short',
                day: 'numeric'
              })}
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default Item;