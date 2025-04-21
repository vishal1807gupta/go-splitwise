import React from "react";
import User from "./User";

const GroupUsers = ({ users}) => {
  if (!users) return null;

  return (
    <div className="space-y-2">
      <div className="flex justify-between items-center mb-2">
        <h2 className="text-lg font-bold text-gray-800">Members</h2>
        <span className="bg-blue-100 text-blue-800 text-xs font-medium px-2 py-0.5 rounded-full">
          {users.length} {users.length === 1 ? 'Member' : 'Members'}
        </span>
      </div>

      {users.length > 0 ? (
        <div className="space-y-1 max-h-72 overflow-y-auto pr-1">
          {users.map((user) => (
            <div
              key={user.id}
              className="bg-white border border-gray-100 rounded-lg hover:bg-gray-50 transition-colors"
            >
              <User user={user} compact={true} />
            </div>
          ))}
        </div>
      ) : (
        <div className="text-center py-6 bg-gray-50 rounded-lg">
          <svg xmlns="http://www.w3.org/2000/svg" className="h-8 w-8 mx-auto text-gray-400 mb-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
          </svg>
          <p className="text-xs text-gray-500">No members yet</p>
        </div>
      )}
    </div>
  );
};

export default GroupUsers;