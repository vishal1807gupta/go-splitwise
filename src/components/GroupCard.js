import React from "react";
import { useNavigate } from "react-router-dom";

const GroupCard = ({ group }) => {
  const navigate = useNavigate();
  
  if (!group) return null;
  
  // Generate consistent colors based on group_id
  const getGroupColor = (id) => {
    const colors = [
      { bg: "bg-blue-50", border: "border-blue-400", accent: "text-blue-600" },
      { bg: "bg-emerald-50", border: "border-emerald-400", accent: "text-emerald-600" },
      { bg: "bg-violet-50", border: "border-violet-400", accent: "text-violet-600" },
      { bg: "bg-amber-50", border: "border-amber-400", accent: "text-amber-600" },
      { bg: "bg-rose-50", border: "border-rose-400", accent: "text-rose-600" },
      { bg: "bg-cyan-50", border: "border-cyan-400", accent: "text-cyan-600" },
      { bg: "bg-fuchsia-50", border: "border-fuchsia-400", accent: "text-fuchsia-600" },
      { bg: "bg-lime-50", border: "border-lime-400", accent: "text-lime-600" },
    ];
    return colors[id % colors.length];
  };
  
  const colorScheme = getGroupColor(group.group_id);
  
  const handleClick = () => {
    console.log("hello");
    navigate(`/groups/${group.group_id}/${group.group_name}`);
  };
  
  return (
    <div 
      onClick={handleClick}
      className={`${colorScheme.bg} border border-gray-200 ${colorScheme.border} border-l-4 rounded-lg shadow-sm hover:shadow-md transition-all duration-200 cursor-pointer h-48 flex flex-col p-5 relative overflow-hidden group`}
    >
      <h3 className="text-xl font-bold text-gray-800 mb-3">{group.group_name}</h3>
      
      {group.description && (
        <p className="text-gray-700 mb-4 line-clamp-3">{group.description}</p>
      )}
      
      <div className="mt-auto flex justify-between items-center">
        {group.member_count !== undefined && (
          <div className="flex items-center text-gray-600">
            <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5 mr-2" viewBox="0 0 20 20" fill="currentColor">
              <path d="M13 6a3 3 0 11-6 0 3 3 0 016 0zM18 8a2 2 0 11-4 0 2 2 0 014 0zM14 15a4 4 0 00-8 0v3h8v-3zM6 8a2 2 0 11-4 0 2 2 0 014 0zM16 18v-3a5.972 5.972 0 00-.75-2.906A3.005 3.005 0 0119 15v3h-3zM4.75 12.094A5.973 5.973 0 004 15v3H1v-3a3 3 0 013.75-2.906z" />
            </svg>
            {group.member_count} {group.member_count === 1 ? 'Member' : 'Members'}
          </div>
        )}
      </div>
      
      {/* Subtle arrow indicator that appears on hover */}
      <div className={`absolute bottom-5 right-5 ${colorScheme.accent} opacity-0 group-hover:opacity-100 transition-opacity duration-200`}>
        <svg xmlns="http://www.w3.org/2000/svg" className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
          <path strokeLinecap="round" strokeLinejoin="round" d="M13 7l5 5m0 0l-5 5m5-5H6" />
        </svg>
      </div>
    </div>
  );
};

export default GroupCard;